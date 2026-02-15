package vcs

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestNewCommit(t *testing.T) {
	metadata := Metadata{
		Author: "Test Author",
	}

	commit := NewCommit(FullBackup, "", "tree123", "Test message", metadata, 1)

	if commit == nil {
		t.Fatal("NewCommit returned nil")
	}

	if commit.Type != FullBackup {
		t.Errorf("Expected type %s, got %s", FullBackup, commit.Type)
	}

	if commit.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", commit.Message)
	}

	if commit.Metadata.Author != "Test Author" {
		t.Errorf("Expected author 'Test Author', got '%s'", commit.Metadata.Author)
	}

	if commit.CommitNumber != 1 {
		t.Errorf("Expected commit number 1, got %d", commit.CommitNumber)
	}

	if commit.ID == "" {
		t.Error("Commit ID should not be empty")
	}
}

func TestCommitTypes(t *testing.T) {
	metadata := Metadata{Author: "Test"}

	fullCommit := NewCommit(FullBackup, "", "tree1", "Full", metadata, 1)
	if !fullCommit.IsFullBackup() {
		t.Error("Expected IsFullBackup to return true")
	}
	if fullCommit.IsDifferential() {
		t.Error("Expected IsDifferential to return false")
	}

	diffCommit := NewCommit(Differential, "parent1", "tree2", "Diff", metadata, 2)
	if diffCommit.IsFullBackup() {
		t.Error("Expected IsFullBackup to return false")
	}
	if !diffCommit.IsDifferential() {
		t.Error("Expected IsDifferential to return true")
	}
}

func TestCommitIDUniqueness(t *testing.T) {
	metadata := Metadata{Author: "Test"}

	commit1 := NewCommit(FullBackup, "", "tree1", "Message 1", metadata, 1)
	commit2 := NewCommit(FullBackup, "", "tree1", "Message 2", metadata, 1)

	if commit1.ID == commit2.ID {
		t.Error("Different commits should have different IDs")
	}
}

func TestCommitWithParent(t *testing.T) {
	metadata := Metadata{Author: "Test"}

	parentCommit := NewCommit(FullBackup, "", "tree1", "Parent", metadata, 1)
	childCommit := NewCommit(Differential, parentCommit.ID, "tree2", "Child", metadata, 2)

	if childCommit.ParentID != parentCommit.ID {
		t.Errorf("Expected parent ID %s, got %s", parentCommit.ID, childCommit.ParentID)
	}
}

func TestInitRepository(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "noms-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tempDir)

	config := RepositoryConfig{
		FullSnapshotInterval: 10,
		Author:               "Test Author",
	}

	repo, err := InitRepository(tempDir, config)
	if err != nil {
		t.Fatalf("InitRepository failed: %v", err)
	}

	// Check repository structure
	nomsDir := filepath.Join(tempDir, NomsDir)
	if _, err := os.Stat(nomsDir); os.IsNotExist(err) {
		t.Error(".noms directory was not created")
	}

	// Check subdirectories
	for _, dir := range []string{CommitsDir, TreesDir, ObjectsDir} {
		dirPath := filepath.Join(nomsDir, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			t.Errorf("%s directory was not created", dir)
		}
	}

	// Check configuration
	if repo.Config.FullSnapshotInterval != 10 {
		t.Errorf("Expected FullSnapshotInterval 10, got %d", repo.Config.FullSnapshotInterval)
	}

	if repo.Config.Author != "Test Author" {
		t.Errorf("Expected author 'Test Author', got '%s'", repo.Config.Author)
	}
}

func TestLoadRepository(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "noms-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tempDir)

	// Initialize repository
	config := RepositoryConfig{
		FullSnapshotInterval: 5,
		Author:               "Test User",
	}

	_, err = InitRepository(tempDir, config)
	if err != nil {
		t.Fatalf("InitRepository failed: %v", err)
	}

	// Load the repository
	repo, err := LoadRepository(tempDir)
	if err != nil {
		t.Fatalf("LoadRepository failed: %v", err)
	}

	// Verify loaded configuration
	if repo.Config.FullSnapshotInterval != 5 {
		t.Errorf("Expected FullSnapshotInterval 5, got %d", repo.Config.FullSnapshotInterval)
	}

	if repo.Config.Author != "Test User" {
		t.Errorf("Expected author 'Test User', got '%s'", repo.Config.Author)
	}
}

func TestShouldCreateFullSnapshot(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "noms-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tempDir)

	config := RepositoryConfig{
		FullSnapshotInterval: 10,
		Author:               "Test",
	}

	repo, err := InitRepository(tempDir, config)
	if err != nil {
		t.Fatalf("InitRepository failed: %v", err)
	}

	// First commit should be full snapshot
	if !repo.ShouldCreateFullSnapshot() {
		t.Error("First commit should be a full snapshot")
	}

	// Simulate commits 2-9 (should not be full snapshots)
	for i := 1; i <= 8; i++ {
		repo.CommitCount = i
		if repo.ShouldCreateFullSnapshot() {
			t.Errorf("Commit %d should not be a full snapshot", i+1)
		}
	}

	// 10th commit should be full snapshot
	repo.CommitCount = 9
	if !repo.ShouldCreateFullSnapshot() {
		t.Error("10th commit should be a full snapshot")
	}
}

func TestBlobStorage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "noms-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tempDir)

	config := RepositoryConfig{
		FullSnapshotInterval: 10,
		Author:               "Test",
	}

	repo, err := InitRepository(tempDir, config)
	if err != nil {
		t.Fatalf("InitRepository failed: %v", err)
	}

	// Test blob storage and retrieval
	testContent := []byte("Hello, World!")
	testHash := "abc123def456"

	// Save blob
	err = repo.SaveBlob(testHash, testContent)
	if err != nil {
		t.Fatalf("SaveBlob failed: %v", err)
	}

	// Check blob exists
	if !repo.BlobExists(testHash) {
		t.Error("BlobExists returned false for saved blob")
	}

	// Load blob
	loadedContent, err := repo.LoadBlob(testHash)
	if err != nil {
		t.Fatalf("LoadBlob failed: %v", err)
	}

	if !bytes.Equal(loadedContent, testContent) {
		t.Errorf("Expected content '%s', got '%s'", testContent, loadedContent)
	}
}
