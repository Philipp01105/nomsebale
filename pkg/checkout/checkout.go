package checkout

import (
	"fmt"
	"noms/pkg/utils"
	"noms/pkg/vcs"
	"os"
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
	
	// Note: For a basic implementation, we just inform the user
	// A full implementation would need to store file contents
	fmt.Println("\nNote: This is a basic version control system.")
	fmt.Println("File restoration requires stored file contents (not yet implemented).")
	fmt.Printf("\nTree state contains %d entries:\n", len(treeState.Entries))
	
	fileCount := 0
	for _, entry := range treeState.Entries {
		if !entry.IsDirectory {
			fileCount++
			fmt.Printf("  %s (hash: %s)\n", entry.Path, utils.TruncateID(entry.Hash))
		}
	}
	
	if fileCount == 0 {
		fmt.Println("  (no files)")
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
	if len(partialID) >= 4 && len(fullID) >= len(partialID) {
		return fullID[:len(partialID)] == partialID
	}
	return false
}

// findCommit finds a commit by full or partial ID
func findCommit(repo *vcs.Repository, partialID string) (string, error) {
	// If no commits exist
	if repo.HEAD == "" {
		return "", fmt.Errorf("no commits in repository")
	}

	// Walk through all commits to find a match
	currentID := repo.HEAD
	for currentID != "" {
		// Check if this commit matches
		if matchesCommitID(currentID, partialID) {
			return currentID, nil
		}

		// Load commit to get parent
		commit, err := repo.LoadCommit(currentID)
		if err != nil {
			return "", fmt.Errorf("error loading commit %s: %w", currentID, err)
		}

		currentID = commit.ParentID
	}

	return "", fmt.Errorf("commit not found: %s", partialID)
}
