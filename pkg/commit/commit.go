package commit

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
	entries, err := scanDirectory(cwd)
	if err != nil {
		fmt.Printf("Error scanning directory: %v\n", err)
		return
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
	fmt.Printf("Created %s commit: %s\n", commit.Type, commit.ID[:8])
	fmt.Printf("Commit number: %d\n", commit.CommitNumber)
	fmt.Printf("Message: %s\n", commit.Message)
	fmt.Printf("Tree state: %s\n", commit.TreeStateID[:8])
	if commit.ParentID != "" {
		fmt.Printf("Parent: %s\n", commit.ParentID[:8])
	}
	fmt.Printf("Timestamp: %s\n", commit.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Files tracked: %d\n", len(entries))
}

// scanDirectory scans a directory and creates tree entries for all files
func scanDirectory(rootPath string) ([]vcs.TreeEntry, error) {
	entries := make([]vcs.TreeEntry, 0)

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .noms directory
		if strings.Contains(path, vcs.NomsDir) {
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
