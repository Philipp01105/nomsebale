package status

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"noms/pkg/vcs"
	"os"
	"path/filepath"
	"strings"
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
	currentEntries, err := scanDirectory(cwd)
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
	fmt.Printf("On commit %s\n", truncateID(repo.HEAD))
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

// truncateID safely truncates an ID to 8 characters for display
func truncateID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

// scanDirectory scans a directory and creates tree entries for all files
func scanDirectory(rootPath string) ([]vcs.TreeEntry, error) {
	entries := make([]vcs.TreeEntry, 0)

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .noms directory
		nomsPath := filepath.Join(rootPath, vcs.NomsDir)
		if path == nomsPath || strings.HasPrefix(path, nomsPath+string(filepath.Separator)) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip the root directory itself
		if path == rootPath {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		// Calculate file hash
		hash := ""
		if !info.IsDir() {
			hash, err = hashFile(path)
			if err != nil {
				return err
			}
		}

		entry := vcs.TreeEntry{
			Path:         relPath,
			Hash:         hash,
			IsDirectory:  info.IsDir(),
			Size:         info.Size(),
			Permissions:  info.Mode().String(),
			ModifiedTime: info.ModTime().Unix(),
		}

		entries = append(entries, entry)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return entries, nil
}

// hashFile computes the SHA256 hash of a file
func hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
