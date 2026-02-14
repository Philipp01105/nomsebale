package status

import (
	"fmt"
	"noms/pkg/utils"
	"noms/pkg/vcs"
	"os"
)

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
		fmt.Println("On initial commit")
		fmt.Printf("\nUntracked files:\n")
		for _, entry := range currentEntries {
			if !entry.IsDirectory {
				fmt.Printf("  %s\n", entry.Path)
			}
		}
		if len(currentEntries) > 0 {
			fmt.Printf("\nTotal: %d untracked files\n", countFiles(currentEntries))
		}
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

	// Display status
	fmt.Printf("On commit %s\n", utils.TruncateID(repo.HEAD))
	fmt.Printf("Commit #%d: %s\n", headCommit.CommitNumber, headCommit.Message)
	
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

	if !hasChanges {
		fmt.Println("\nWorking tree clean")
	} else {
		fmt.Printf("\nChanges: %d modified, %d added, %d deleted\n",
			len(delta.Modified),
			countFiles(delta.Added),
			len(delta.Deleted))
	}
}

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
