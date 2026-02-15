package status

import (
	"fmt"
	"os"

	"github.com/Philipp01105/nomsebale/pkg/utils"
	"github.com/Philipp01105/nomsebale/pkg/vcs"
)

// countFiles counts non-directory entries
func countFiles(entries []vcs.TreeEntry) int {
	count := 0
	for _, entry := range entries {
		if !entry.IsDirectory {
			count++
		}
	}
	return count
}

// displayNoCommitsStatus displays status when there are no commits yet
func displayNoCommitsStatus(repo *vcs.Repository, currentEntries []vcs.TreeEntry) {
	currentBranch, err := repo.GetCurrentBranch()
	if err == nil && currentBranch != "" {
		fmt.Printf("On branch %s\n", currentBranch)
	}
	fmt.Println("No commits yet")
	fmt.Printf("\nUntracked files:\n")
	for _, entry := range currentEntries {
		if !entry.IsDirectory {
			fmt.Printf("  %s\n", entry.Path)
		}
	}
	if len(currentEntries) > 0 {
		fmt.Printf("\nTotal: %d untracked files\n", countFiles(currentEntries))
	}
}

// displayBranchInfo displays current branch or detached HEAD info
func displayBranchInfo(repo *vcs.Repository, currentBranch string) {
	if currentBranch != "" {
		fmt.Printf("On branch %s\n", currentBranch)
	} else {
		fmt.Printf("HEAD detached at %s\n", utils.TruncateID(repo.HEAD))
	}
}

// displayChanges displays modified, added, and deleted files
func displayChanges(delta *vcs.Delta) bool {
	hasChanges := false

	if len(delta.Modified) > 0 {
		hasChanges = true
		fmt.Printf("\nModified files:\n")
		for _, entry := range delta.Modified {
			fmt.Printf("  %s\n", entry.Path)
		}
	}

	if len(delta.Added) > 0 {
		hasChanges = true
		fmt.Printf("\nNew files:\n")
		for _, entry := range delta.Added {
			if !entry.IsDirectory {
				fmt.Printf("  %s\n", entry.Path)
			}
		}
	}

	if len(delta.Deleted) > 0 {
		hasChanges = true
		fmt.Printf("\nDeleted files:\n")
		for _, path := range delta.Deleted {
			fmt.Printf("  %s\n", path)
		}
	}

	return hasChanges
}

// Status shows the current working tree status
func Status() {
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

	// Scan current directory
	currentEntries, err := utils.ScanDirectory(cwd)
	if err != nil {
		fmt.Printf("Error scanning directory: %v\n", err)
		return
	}

	// If there are no commits yet
	if repo.HEAD == "" {
		displayNoCommitsStatus(repo, currentEntries)
		return
	}

	// Load HEAD commit
	headCommit, err := repo.LoadCommit(repo.HEAD)
	if err != nil {
		fmt.Printf("Error loading HEAD commit: %v\n", err)
		return
	}

	// Load the tree state
	treeState, err := repo.LoadTreeState(headCommit.TreeStateID)
	if err != nil {
		fmt.Printf("Error loading tree state: %v\n", err)
		return
	}

	// Compute delta
	currentTreeState := vcs.NewTreeState(currentEntries)
	delta := vcs.ComputeDelta(treeState, currentTreeState)

	// Get current branch
	currentBranch, err := repo.GetCurrentBranch()

	// Display status
	displayBranchInfo(repo, currentBranch)
	fmt.Printf("Latest commit: %s\n", utils.TruncateID(repo.HEAD))
	fmt.Printf("Commit #%d: %s\n", headCommit.CommitNumber, headCommit.Message)

	hasChanges := displayChanges(delta)

	if !hasChanges {
		fmt.Println("\nWorking tree clean")
	} else {
		fmt.Printf("\nChanges: %d modified, %d added, %d deleted\n",
			len(delta.Modified),
			countFiles(delta.Added),
			len(delta.Deleted))
	}
}
