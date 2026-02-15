package checkout

import (
	"fmt"
	"noms/pkg/utils"
	"noms/pkg/vcs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Checkout restores files from a specific commit
func Checkout(commitID string) {
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

	// Find the commit (allow partial commit IDs)
	fullCommitID, err := findCommit(repo, commitID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Load the commit
	commit, err := repo.LoadCommit(fullCommitID)
	if err != nil {
		fmt.Printf("Error loading commit: %v\n", err)
		return
	}

	// Load the tree state
	treeState, err := repo.LoadTreeState(commit.TreeStateID)
	if err != nil {
		fmt.Printf("Error loading tree state: %v\n", err)
		return
	}

	// Restore files from tree state
	fmt.Printf("Checking out commit %s\n", utils.TruncateID(fullCommitID))
	fmt.Printf("Commit #%d: %s\n", commit.CommitNumber, commit.Message)
	fmt.Printf("\nRestoring %d entries...\n", len(treeState.Entries))
	
	restoredCount := 0
	skippedCount := 0
	failedCount := 0
	
	// Get current files to track what should be deleted
	currentEntries, err := utils.ScanDirectory(cwd)
	if err != nil {
		fmt.Printf("Warning: failed to scan current directory: %v\n", err)
	}
	
	currentFiles := make(map[string]bool)
	for _, entry := range currentEntries {
		if !entry.IsDirectory {
			currentFiles[entry.Path] = true
		}
	}
	
	// Track files in the tree state
	treeFiles := make(map[string]bool)
	
	// Restore files from the tree state
	for _, entry := range treeState.Entries {
		if entry.IsDirectory {
			// Create directory
			dirPath := utils.JoinPath(cwd, entry.Path)
			if err := os.MkdirAll(dirPath, 0755); err != nil {
				fmt.Printf("  Failed to create directory %s: %v\n", entry.Path, err)
				failedCount++
			}
			continue
		}
		
		treeFiles[entry.Path] = true
		
		// Check if file already exists with same hash
		filePath := utils.JoinPath(cwd, entry.Path)
		if _, err := os.Stat(filePath); err == nil {
			currentHash, err := utils.HashFile(filePath)
			if err == nil && currentHash == entry.Hash {
				skippedCount++
				continue
			}
		}
		
		// Load blob content
		content, err := repo.LoadBlob(entry.Hash)
		if err != nil {
			fmt.Printf("  Failed to load content for %s: %v\n", entry.Path, err)
			failedCount++
			continue
		}
		
		// Ensure parent directory exists
		parentDir := filepath.Dir(filePath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			fmt.Printf("  Failed to create parent directory for %s: %v\n", entry.Path, err)
			failedCount++
			continue
		}
		
		// Write file
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			fmt.Printf("  Failed to write file %s: %v\n", entry.Path, err)
			failedCount++
			continue
		}
		
		// Try to restore permissions (best effort)
		if mode, err := parsePermissions(entry.Permissions); err == nil {
			if err := os.Chmod(filePath, mode); err != nil {
				// Log permission restoration failures for debugging
				fmt.Printf("  Warning: failed to restore permissions for %s: %v\n", entry.Path, err)
			}
		}
		
		restoredCount++
	}
	
	// Delete files that are not in the tree state
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
	
	fmt.Printf("\nCheckout complete:\n")
	fmt.Printf("  Restored: %d files\n", restoredCount)
	if skippedCount > 0 {
		fmt.Printf("  Skipped (unchanged): %d files\n", skippedCount)
	}
	if deletedCount > 0 {
		fmt.Printf("  Deleted: %d files\n", deletedCount)
	}
	if failedCount > 0 {
		fmt.Printf("  Failed: %d files\n", failedCount)
	}

	// Update HEAD to point to this commit
	if err := repo.UpdateHEAD(fullCommitID); err != nil {
		fmt.Printf("Error updating HEAD: %v\n", err)
		return
	}

	// Save config with updated HEAD
	if err := repo.SaveConfig(); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		return
	}

	fmt.Printf("\nHEAD is now at %s\n", utils.TruncateID(fullCommitID))
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
// Handles strings like "-rw-r--r--" or "0644"
func parsePermissions(permStr string) (os.FileMode, error) {
	// If it looks like octal (starts with digit), try parsing as octal
	if len(permStr) > 0 && permStr[0] >= '0' && permStr[0] <= '9' {
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
			mode |= 0400
		}
		if permStr[2] == 'w' {
			mode |= 0200
		}
		if permStr[3] == 'x' {
			mode |= 0100
		}
		
		// Group permissions
		if permStr[4] == 'r' {
			mode |= 0040
		}
		if permStr[5] == 'w' {
			mode |= 0020
		}
		if permStr[6] == 'x' {
			mode |= 0010
		}
		
		// Other permissions
		if permStr[7] == 'r' {
			mode |= 0004
		}
		if permStr[8] == 'w' {
			mode |= 0002
		}
		if permStr[9] == 'x' {
			mode |= 0001
		}
		
		return mode, nil
	}
	
	return 0, fmt.Errorf("invalid permission string: %s", permStr)
}
