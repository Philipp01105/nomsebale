package checkout

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Philipp01105/nomsebale/pkg/utils"
	"github.com/Philipp01105/nomsebale/pkg/vcs"
)

// Checkout restores files from a specific commit or switches to a branch
func Checkout(ref string) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return
	}

	// Load repository
	repo, err := vcs.LoadRepository(cwd)
	if err != nil {
		fmt.Printf("Error: not a noms repository. Run 'noms init' first.\n")
		return
	}

	// Resolve the reference to a commit ID
	commitID, isBranch, branchName, err := resolveCheckoutRef(repo, ref)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Handle empty commit case
	if commitID == "" {
		handleEmptyCommit(repo, isBranch, branchName)
		return
	}

	// Load the commit and tree state
	commit, treeState, err := loadCommitAndTree(repo, commitID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print checkout info
	printCheckoutInfo(isBranch, branchName, commitID, commit)
	fmt.Printf("\nRestoring %d entries...\n", len(treeState.Entries))

	// Restore files from tree state
	stats := restoreTreeState(repo, cwd, treeState)

	// Delete files that are not in the tree state
	stats.deletedCount = deleteUnwantedFiles(cwd, stats.treeFiles, stats.currentFiles)

	// Print summary
	printCheckoutSummary(stats)

	// Update HEAD appropriately
	if err := updateRepositoryHead(repo, isBranch, branchName, commitID); err != nil {
		fmt.Printf("Error updating HEAD: %v\n", err)
		return
	}
}

type checkoutStats struct {
	restoredCount int
	skippedCount  int
	failedCount   int
	deletedCount  int
	treeFiles     map[string]bool
	currentFiles  map[string]bool
}

// resolveCheckoutRef resolves a reference (branch or commit) to a commit ID
func resolveCheckoutRef(repo *vcs.Repository, ref string) (commitID string, isBranch bool, branchName string, err error) {
	// Check if ref is a branch name
	if repo.BranchExists(ref) {
		branch, e := repo.GetBranch(ref)
		if e != nil {
			err = fmt.Errorf("loading branch: %w", e)
			return
		}
		commitID = branch.CommitID
		isBranch = true
		branchName = ref
		return
	}

	// Try to find commit by ID (allow partial commit IDs)
	fullCommitID, e := findCommit(repo, ref)
	if e != nil {
		err = e
		return
	}
	commitID = fullCommitID
	isBranch = false
	return
}

// handleEmptyCommit handles the case when checking out a branch with no commits
func handleEmptyCommit(repo *vcs.Repository, isBranch bool, branchName string) {
	if isBranch {
		// Switch to branch even if it has no commits yet
		if err := repo.SetCurrentBranch(branchName); err != nil {
			fmt.Printf("Error switching to branch: %v\n", err)
			return
		}
		fmt.Printf("Switched to branch '%s'\n", branchName)
		fmt.Println("No commits on this branch yet")
		return
	}
	fmt.Println("Error: cannot checkout empty reference")
}

// loadCommitAndTree loads a commit and its associated tree state
func loadCommitAndTree(repo *vcs.Repository, commitID string) (*vcs.Commit, *vcs.TreeState, error) {
	commit, err := repo.LoadCommit(commitID)
	if err != nil {
		return nil, nil, fmt.Errorf("loading commit: %w", err)
	}

	treeState, err := repo.LoadTreeState(commit.TreeStateID)
	if err != nil {
		return nil, nil, fmt.Errorf("loading tree state: %w", err)
	}

	return commit, treeState, nil
}

// printCheckoutInfo prints information about the checkout operation
func printCheckoutInfo(isBranch bool, branchName, commitID string, commit *vcs.Commit) {
	if isBranch {
		fmt.Printf("Switching to branch '%s'\n", branchName)
	} else {
		fmt.Printf("Checking out commit %s\n", utils.TruncateID(commitID))
		fmt.Printf("Note: switching to detached HEAD state\n")
	}
	fmt.Printf("Commit #%d: %s\n", commit.CommitNumber, commit.Message)
}

// restoreTreeState restores files from the tree state
func restoreTreeState(repo *vcs.Repository, cwd string, treeState *vcs.TreeState) checkoutStats {
	stats := checkoutStats{
		treeFiles:    make(map[string]bool),
		currentFiles: make(map[string]bool),
	}

	// Get current files to track what should be deleted
	currentEntries, err := utils.ScanDirectory(cwd)
	if err != nil {
		fmt.Printf("Warning: failed to scan current directory: %v\n", err)
	}

	for _, entry := range currentEntries {
		if !entry.IsDirectory {
			stats.currentFiles[entry.Path] = true
		}
	}

	// Restore files from the tree state
	for _, entry := range treeState.Entries {
		if entry.IsDirectory {
			// Create directory
			dirPath := utils.JoinPath(cwd, entry.Path)
			if err := os.MkdirAll(dirPath, 0o755); err != nil {
				fmt.Printf("  Failed to create directory %s: %v\n", entry.Path, err)
				stats.failedCount++
			}
			continue
		}

		stats.treeFiles[entry.Path] = true

		// Check if file already exists with same hash
		filePath := utils.JoinPath(cwd, entry.Path)
		if _, err := os.Stat(filePath); err == nil {
			currentHash, err := utils.HashFile(filePath)
			if err == nil && currentHash == entry.Hash {
				stats.skippedCount++
				continue
			}
		}

		// Load blob content
		content, err := repo.LoadBlob(entry.Hash)
		if err != nil {
			fmt.Printf("  Failed to load content for %s: %v\n", entry.Path, err)
			stats.failedCount++
			continue
		}

		// Ensure parent directory exists
		parentDir := filepath.Dir(filePath)
		if err := os.MkdirAll(parentDir, 0o755); err != nil {
			fmt.Printf("  Failed to create parent directory for %s: %v\n", entry.Path, err)
			stats.failedCount++
			continue
		}

		// Write file
		if err := os.WriteFile(filePath, content, 0o644); err != nil {
			fmt.Printf("  Failed to write file %s: %v\n", entry.Path, err)
			stats.failedCount++
			continue
		}

		// Try to restore permissions (best effort)
		if mode, err := parsePermissions(entry.Permissions); err == nil {
			if err := os.Chmod(filePath, mode); err != nil {
				// Log permission restoration failures for debugging
				fmt.Printf("  Warning: failed to restore permissions for %s: %v\n", entry.Path, err)
			}
		}

		stats.restoredCount++
	}

	return stats
}

// deleteUnwantedFiles deletes files that are not in the tree state
func deleteUnwantedFiles(cwd string, treeFiles, currentFiles map[string]bool) int {
	deletedCount := 0
	for filePath := range currentFiles {
		if !treeFiles[filePath] {
			absPath := utils.JoinPath(cwd, filePath)
			if err := os.Remove(absPath); err != nil {
				fmt.Printf("  Failed to delete %s: %v\n", filePath, err)
			} else {
				deletedCount++
			}
		}
	}
	return deletedCount
}

// printCheckoutSummary prints a summary of the checkout operation
func printCheckoutSummary(stats checkoutStats) {
	fmt.Printf("\nCheckout complete:\n")
	fmt.Printf("  Restored: %d files\n", stats.restoredCount)
	if stats.skippedCount > 0 {
		fmt.Printf("  Skipped (unchanged): %d files\n", stats.skippedCount)
	}
	if stats.deletedCount > 0 {
		fmt.Printf("  Deleted: %d files\n", stats.deletedCount)
	}
	if stats.failedCount > 0 {
		fmt.Printf("  Failed: %d files\n", stats.failedCount)
	}
}

// updateRepositoryHead updates HEAD to point to the checked out ref
func updateRepositoryHead(repo *vcs.Repository, isBranch bool, branchName, commitID string) error {
	if isBranch {
		// Set HEAD to point to the branch
		if err := repo.SetCurrentBranch(branchName); err != nil {
			return fmt.Errorf("setting branch: %w", err)
		}
		fmt.Printf("\nSwitched to branch '%s'\n", branchName)
	} else {
		// Detached HEAD - point directly to commit
		if err := repo.UpdateHEAD(commitID); err != nil {
			return fmt.Errorf("updating HEAD: %w", err)
		}
		fmt.Printf("\nHEAD is now at %s (detached)\n", utils.TruncateID(commitID))
	}

	// Save config with updated HEAD
	if err := repo.SaveConfig(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	return nil
}

// matchesCommitID checks if a commit ID matches a partial ID
func matchesCommitID(fullID, partialID string) bool {
	if fullID == partialID {
		return true
	}
	if len(partialID) >= 4 && len(partialID) <= len(fullID) {
		return fullID[:len(partialID)] == partialID
	}
	return false
}

// findCommit finds a commit by full or partial ID
func findCommit(repo *vcs.Repository, partialID string) (string, error) {
	// Get all commit files in the commits directory
	commitsDir := filepath.Join(repo.Path, vcs.NomsDir, vcs.CommitsDir)
	entries, err := os.ReadDir(commitsDir)
	if err != nil {
		return "", fmt.Errorf("error reading commits directory: %w", err)
	}

	if len(entries) == 0 {
		return "", fmt.Errorf("no commits in repository")
	}

	// Search through all commits
	var matches []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Extract commit ID from filename (remove .json extension)
		filename := entry.Name()
		if !strings.HasSuffix(filename, ".json") {
			continue
		}
		commitID := filename[:len(filename)-5]

		// Check if this commit matches
		if matchesCommitID(commitID, partialID) {
			matches = append(matches, commitID)
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("commit not found: %s", partialID)
	}

	if len(matches) > 1 {
		return "", fmt.Errorf("ambiguous commit ID %s, matches: %d commits", partialID, len(matches))
	}

	return matches[0], nil
}

// parsePermissions converts a permission string to os.FileMode
// Handles strings like "-rw-r--r--" or "0o644"
func parsePermissions(permStr string) (os.FileMode, error) {
	// If it looks like octal (starts with digit), try parsing as octal
	if permStr != "" && permStr[0] >= '0' && permStr[0] <= '9' {
		perm, err := strconv.ParseUint(permStr, 8, 32)
		if err != nil {
			return 0, fmt.Errorf("invalid octal permission: %w", err)
		}
		return os.FileMode(perm), nil
	}

	// Try to parse symbolic notation like "-rw-r--r--"
	if len(permStr) >= 10 {
		var mode os.FileMode

		// Owner permissions
		if permStr[1] == 'r' {
			mode |= 0o400
		}
		if permStr[2] == 'w' {
			mode |= 0o200
		}
		if permStr[3] == 'x' {
			mode |= 0o100
		}

		// Group permissions
		if permStr[4] == 'r' {
			mode |= 0o040
		}
		if permStr[5] == 'w' {
			mode |= 0o020
		}
		if permStr[6] == 'x' {
			mode |= 0o010
		}

		// Other permissions
		if permStr[7] == 'r' {
			mode |= 0o004
		}
		if permStr[8] == 'w' {
			mode |= 0o002
		}
		if permStr[9] == 'x' {
			mode |= 0o001
		}

		return mode, nil
	}

	return 0, fmt.Errorf("invalid permission string: %s", permStr)
}
