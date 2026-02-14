package log

import (
	"fmt"
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

	// Walk through commit history
	currentID := repo.HEAD
	for currentID != "" {
		commit, err := repo.LoadCommit(currentID)
		if err != nil {
			fmt.Printf("Error loading commit %s: %v\n", truncateID(currentID), err)
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

// truncateID safely truncates an ID to 8 characters for display
func truncateID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
