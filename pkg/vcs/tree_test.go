package vcs

import (
	"testing"
)

func TestNewTreeState(t *testing.T) {
	entries := []TreeEntry{
		{Path: "file1.txt", Hash: "hash1", IsDirectory: false},
		{Path: "file2.txt", Hash: "hash2", IsDirectory: false},
	}

	treeState := NewTreeState(entries)

	if treeState == nil {
		t.Fatal("NewTreeState returned nil")
	}

	if len(treeState.Entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(treeState.Entries))
	}

	if treeState.ID == "" {
		t.Error("TreeState ID should not be empty")
	}
}

func TestTreeStateIDUniqueness(t *testing.T) {
	entries1 := []TreeEntry{
		{Path: "file1.txt", Hash: "hash1", IsDirectory: false},
	}

	entries2 := []TreeEntry{
		{Path: "file2.txt", Hash: "hash2", IsDirectory: false},
	}

	tree1 := NewTreeState(entries1)
	tree2 := NewTreeState(entries2)

	if tree1.ID == tree2.ID {
		t.Error("Different tree states should have different IDs")
	}
}

func TestTreeStateIDConsistency(t *testing.T) {
	entries := []TreeEntry{
		{Path: "file1.txt", Hash: "hash1", IsDirectory: false},
		{Path: "file2.txt", Hash: "hash2", IsDirectory: false},
	}

	tree1 := NewTreeState(entries)
	tree2 := NewTreeState(entries)

	if tree1.ID != tree2.ID {
		t.Error("Same tree states should have the same ID")
	}
}

func TestComputeDelta(t *testing.T) {
	oldEntries := []TreeEntry{
		{Path: "file1.txt", Hash: "hash1", IsDirectory: false},
		{Path: "file2.txt", Hash: "hash2", IsDirectory: false},
		{Path: "file3.txt", Hash: "hash3", IsDirectory: false},
	}

	newEntries := []TreeEntry{
		{Path: "file1.txt", Hash: "hash1", IsDirectory: false},          // Unchanged
		{Path: "file2.txt", Hash: "hash2-modified", IsDirectory: false}, // Modified
		{Path: "file4.txt", Hash: "hash4", IsDirectory: false},          // Added
	}

	oldState := NewTreeState(oldEntries)
	newState := NewTreeState(newEntries)

	delta := ComputeDelta(oldState, newState)

	// Check added files
	if len(delta.Added) != 1 {
		t.Errorf("Expected 1 added file, got %d", len(delta.Added))
	}
	if len(delta.Added) > 0 && delta.Added[0].Path != "file4.txt" {
		t.Errorf("Expected added file 'file4.txt', got '%s'", delta.Added[0].Path)
	}

	// Check modified files
	if len(delta.Modified) != 1 {
		t.Errorf("Expected 1 modified file, got %d", len(delta.Modified))
	}
	if len(delta.Modified) > 0 && delta.Modified[0].Path != "file2.txt" {
		t.Errorf("Expected modified file 'file2.txt', got '%s'", delta.Modified[0].Path)
	}

	// Check deleted files
	if len(delta.Deleted) != 1 {
		t.Errorf("Expected 1 deleted file, got %d", len(delta.Deleted))
	}
	if len(delta.Deleted) > 0 && delta.Deleted[0] != "file3.txt" {
		t.Errorf("Expected deleted file 'file3.txt', got '%s'", delta.Deleted[0])
	}
}

func TestComputeDeltaNoChanges(t *testing.T) {
	entries := []TreeEntry{
		{Path: "file1.txt", Hash: "hash1", IsDirectory: false},
		{Path: "file2.txt", Hash: "hash2", IsDirectory: false},
	}

	oldState := NewTreeState(entries)
	newState := NewTreeState(entries)

	delta := ComputeDelta(oldState, newState)

	if len(delta.Added) != 0 {
		t.Errorf("Expected 0 added files, got %d", len(delta.Added))
	}

	if len(delta.Modified) != 0 {
		t.Errorf("Expected 0 modified files, got %d", len(delta.Modified))
	}

	if len(delta.Deleted) != 0 {
		t.Errorf("Expected 0 deleted files, got %d", len(delta.Deleted))
	}
}

func TestComputeDeltaAllNew(t *testing.T) {
	oldEntries := []TreeEntry{}

	newEntries := []TreeEntry{
		{Path: "file1.txt", Hash: "hash1", IsDirectory: false},
		{Path: "file2.txt", Hash: "hash2", IsDirectory: false},
	}

	oldState := NewTreeState(oldEntries)
	newState := NewTreeState(newEntries)

	delta := ComputeDelta(oldState, newState)

	if len(delta.Added) != 2 {
		t.Errorf("Expected 2 added files, got %d", len(delta.Added))
	}

	if len(delta.Modified) != 0 {
		t.Errorf("Expected 0 modified files, got %d", len(delta.Modified))
	}

	if len(delta.Deleted) != 0 {
		t.Errorf("Expected 0 deleted files, got %d", len(delta.Deleted))
	}
}

func TestComputeDeltaAllDeleted(t *testing.T) {
	oldEntries := []TreeEntry{
		{Path: "file1.txt", Hash: "hash1", IsDirectory: false},
		{Path: "file2.txt", Hash: "hash2", IsDirectory: false},
	}

	newEntries := []TreeEntry{}

	oldState := NewTreeState(oldEntries)
	newState := NewTreeState(newEntries)

	delta := ComputeDelta(oldState, newState)

	if len(delta.Added) != 0 {
		t.Errorf("Expected 0 added files, got %d", len(delta.Added))
	}

	if len(delta.Modified) != 0 {
		t.Errorf("Expected 0 modified files, got %d", len(delta.Modified))
	}

	if len(delta.Deleted) != 2 {
		t.Errorf("Expected 2 deleted files, got %d", len(delta.Deleted))
	}
}
