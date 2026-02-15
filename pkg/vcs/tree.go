package vcs

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// TreeState represents the state of the file tree at a point in time
type TreeState struct {
	ID      string      `json:"id"`
	Entries []TreeEntry `json:"entries"`
}

// TreeEntry represents a file or directory in the tree
type TreeEntry struct {
	Path         string `json:"path"`
	Hash         string `json:"hash"`
	IsDirectory  bool   `json:"is_directory"`
	Size         int64  `json:"size"`
	Permissions  string `json:"permissions"`
	ModifiedTime int64  `json:"modified_time"`
}

// NewTreeState creates a new tree state from entries
func NewTreeState(entries []TreeEntry) *TreeState {
	ts := &TreeState{
		Entries: entries,
	}
	ts.ID = ts.generateID()
	return ts
}

// generateID generates a unique hash for the tree state
// Hash = SHA256(concatenated entry hashes and paths)
func (ts *TreeState) generateID() string {
	var builder strings.Builder
	for _, entry := range ts.Entries {
		builder.WriteString(entry.Path)
		builder.WriteString(":")
		builder.WriteString(entry.Hash)
		builder.WriteString(";")
	}
	hash := sha256.Sum256([]byte(builder.String()))
	return hex.EncodeToString(hash[:])
}

// Delta represents changes between two tree states
type Delta struct {
	Added    []TreeEntry `json:"added"`
	Modified []TreeEntry `json:"modified"`
	Deleted  []string    `json:"deleted"` // paths
}

// ComputeDelta computes the differences between two tree states
func ComputeDelta(oldState, newState *TreeState) *Delta {
	delta := &Delta{
		Added:    make([]TreeEntry, 0),
		Modified: make([]TreeEntry, 0),
		Deleted:  make([]string, 0),
	}

	// Create maps for efficient lookup
	oldMap := make(map[string]TreeEntry)
	for _, entry := range oldState.Entries {
		oldMap[entry.Path] = entry
	}

	newMap := make(map[string]TreeEntry)
	for _, entry := range newState.Entries {
		newMap[entry.Path] = entry
	}

	// Find added and modified files
	for path, newEntry := range newMap {
		if oldEntry, exists := oldMap[path]; exists {
			// File exists in both - check if modified
			if oldEntry.Hash != newEntry.Hash {
				delta.Modified = append(delta.Modified, newEntry)
			}
		} else {
			// File is new
			delta.Added = append(delta.Added, newEntry)
		}
	}

	// Find deleted files
	for path := range oldMap {
		if _, exists := newMap[path]; !exists {
			delta.Deleted = append(delta.Deleted, path)
		}
	}

	return delta
}
