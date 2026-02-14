package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"noms/pkg/vcs"
	"os"
	"path/filepath"
	"strings"
)

// TruncateID safely truncates an ID to 8 characters for display
func TruncateID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

// ScanDirectory scans a directory and creates tree entries for all files
func ScanDirectory(rootPath string) ([]vcs.TreeEntry, error) {
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
			hash, err = HashFile(path)
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

// HashFile computes the SHA256 hash of a file
func HashFile(path string) (string, error) {
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
