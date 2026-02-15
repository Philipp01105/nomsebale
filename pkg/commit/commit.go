package commit

import (
	"fmt"
	"noms/pkg/utils"
	"noms/pkg/vcs"
	"os"
)

// Commit creates a new commit in the repository
func Commit(message string) {
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

	// Scan current directory for files
	entries, err := utils.ScanDirectory(cwd)
	if err != nil {
		fmt.Printf("Error scanning directory: %v\n", err)
		return
	}

	// Store file contents for all files
	fileCount := 0
	for _, entry := range entries {
		if !entry.IsDirectory && entry.Hash != "" {
			// Check if blob already exists to avoid duplicate storage
			if !repo.BlobExists(entry.Hash) {
				// Read file content
				filePath := utils.GetAbsolutePath(cwd, entry.Path)
				content, err := os.ReadFile(filePath)
				if err != nil {
					fmt.Printf("Warning: failed to read file %s: %v\n", entry.Path, err)
					continue
				}
				
				// Store the blob
				if err := repo.SaveBlob(entry.Hash, content); err != nil {
					fmt.Printf("Warning: failed to store file %s: %v\n", entry.Path, err)
					continue
				}
			}
			fileCount++
		}
	}

	// Create tree state
	treeState := vcs.NewTreeState(entries)

	// Create commit
	commit, err := repo.CreateCommit(treeState, message)
	if err != nil {
		fmt.Printf("Error creating commit: %v\n", err)
		return
	}

	// Print commit information
	fmt.Printf("Created %s commit: %s\n", commit.Type, utils.TruncateID(commit.ID))
	fmt.Printf("Commit number: %d\n", commit.CommitNumber)
	fmt.Printf("Message: %s\n", commit.Message)
	fmt.Printf("Tree state: %s\n", utils.TruncateID(commit.TreeStateID))
	if commit.ParentID != "" {
		fmt.Printf("Parent: %s\n", utils.TruncateID(commit.ParentID))
	}
	fmt.Printf("Timestamp: %s\n", commit.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Files tracked: %d\n", fileCount)
}
