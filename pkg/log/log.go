package log

import (
	"fmt"
	"noms/pkg/utils"
	"noms/pkg/vcs"
	"os"
)

// Log displays the commit history
func Log() {
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

	// Check if there are any commits
	if repo.HEAD == "" {
		fmt.Println("No commits yet")
		return
	}

	// Get current branch
	currentBranch, _ := repo.GetCurrentBranch()
	if currentBranch != "" {
		fmt.Printf("Branch: %s\n\n", currentBranch)
	} else {
		fmt.Printf("HEAD detached at %s\n\n", utils.TruncateID(repo.HEAD))
	}

	// Walk through commit history
	currentID := repo.HEAD
	for currentID != "" {
		commit, err := repo.LoadCommit(currentID)
		if err != nil {
			fmt.Printf("Error loading commit %s: %v\n", utils.TruncateID(currentID), err)
			return
		}

		// Print commit information
		fmt.Printf("commit %s\n", commit.ID)
		fmt.Printf("Type: %s\n", commit.Type)
		fmt.Printf("Author: %s\n", commit.Metadata.Author)
		fmt.Printf("Date: %s\n", commit.Timestamp.Format("Mon Jan 2 15:04:05 2006"))
		fmt.Printf("Commit #%d\n", commit.CommitNumber)
		fmt.Printf("\n    %s\n\n", commit.Message)

		// Move to parent commit
		currentID = commit.ParentID
	}
}
