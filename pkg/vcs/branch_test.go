package vcs

import (
	"os"
	"path/filepath"
	"testing"
)

const testBranchName = "test-branch"

func TestCreateBranch(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "noms-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	config := RepositoryConfig{
		FullSnapshotInterval: 10,
		Author:               "Test",
	}

	repo, err := InitRepository(tempDir, config)
	if err != nil {
		t.Fatalf("InitRepository failed: %v", err)
	}

	// Create a branch
	err = repo.CreateBranch(testBranchName, "")
	if err != nil {
		t.Fatalf("CreateBranch failed: %v", err)
	}

	// Check if branch exists
	if !repo.BranchExists(testBranchName) {
		t.Error("Branch should exist after creation")
	}

	// Try to create the same branch again (should fail)
	err = repo.CreateBranch(testBranchName, "")
	if err == nil {
		t.Error("Creating duplicate branch should fail")
	}
}

func TestBranchNameValidation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "noms-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	config := RepositoryConfig{
		FullSnapshotInterval: 10,
		Author:               "Test",
	}

	repo, err := InitRepository(tempDir, config)
	if err != nil {
		t.Fatalf("InitRepository failed: %v", err)
	}

	// Test invalid branch names
	invalidNames := []string{
		"",            // Empty
		"branch name", // Space
		"branch/name", // Slash
		"branch^name", // Caret
		"branch:name", // Colon
		"branch*name", // Asterisk
		"branch?name", // Question mark
	}

	for _, name := range invalidNames {
		err := repo.CreateBranch(name, "")
		if err == nil {
			t.Errorf("Branch name '%s' should be invalid", name)
		}
	}

	// Test valid branch name
	err = repo.CreateBranch("valid-branch-name", "")
	if err != nil {
		t.Errorf("Valid branch name should be accepted: %v", err)
	}
}

func TestGetBranch(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "noms-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	config := RepositoryConfig{
		FullSnapshotInterval: 10,
		Author:               "Test",
	}

	repo, err := InitRepository(tempDir, config)
	if err != nil {
		t.Fatalf("InitRepository failed: %v", err)
	}

	// Create a branch
	branchName := testBranchName
	err = repo.CreateBranch(branchName, "")
	if err != nil {
		t.Fatalf("CreateBranch failed: %v", err)
	}

	// Get the branch
	branch, err := repo.GetBranch(branchName)
	if err != nil {
		t.Fatalf("GetBranch failed: %v", err)
	}

	if branch.Name != branchName {
		t.Errorf("Expected branch name '%s', got '%s'", branchName, branch.Name)
	}
}

func TestListBranches(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "noms-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	config := RepositoryConfig{
		FullSnapshotInterval: 10,
		Author:               "Test",
	}

	repo, err := InitRepository(tempDir, config)
	if err != nil {
		t.Fatalf("InitRepository failed: %v", err)
	}

	// Create multiple branches
	branches := []string{"branch1", "branch2", "branch3"}
	for _, name := range branches {
		err := repo.CreateBranch(name, "")
		if err != nil {
			t.Fatalf("CreateBranch failed for '%s': %v", name, err)
		}
	}

	// List all branches
	listedBranches, err := repo.ListBranches()
	if err != nil {
		t.Fatalf("ListBranches failed: %v", err)
	}

	if len(listedBranches) != len(branches) {
		t.Errorf("Expected %d branches, got %d", len(branches), len(listedBranches))
	}

	// Verify all branches are listed
	branchMap := make(map[string]bool)
	for _, b := range listedBranches {
		branchMap[b.Name] = true
	}

	for _, name := range branches {
		if !branchMap[name] {
			t.Errorf("Branch '%s' not found in list", name)
		}
	}
}

func TestDeleteBranch(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "noms-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	config := RepositoryConfig{
		FullSnapshotInterval: 10,
		Author:               "Test",
	}

	repo, err := InitRepository(tempDir, config)
	if err != nil {
		t.Fatalf("InitRepository failed: %v", err)
	}

	// Create a branch
	branchName := testBranchName
	err = repo.CreateBranch(branchName, "")
	if err != nil {
		t.Fatalf("CreateBranch failed: %v", err)
	}

	// Create and set another branch as current
	err = repo.CreateBranch("other-branch", "")
	if err != nil {
		t.Fatalf("CreateBranch failed: %v", err)
	}
	err = repo.SetCurrentBranch("other-branch")
	if err != nil {
		t.Fatalf("SetCurrentBranch failed: %v", err)
	}

	// Delete the first branch
	err = repo.DeleteBranch(branchName)
	if err != nil {
		t.Fatalf("DeleteBranch failed: %v", err)
	}

	// Verify branch is deleted
	if repo.BranchExists(branchName) {
		t.Error("Branch should not exist after deletion")
	}

	// Verify branch file is removed
	branchPath := filepath.Join(tempDir, NomsDir, RefsDir, HeadsDir, branchName+".json")
	if _, err := os.Stat(branchPath); !os.IsNotExist(err) {
		t.Error("Branch file should be removed")
	}
}

func TestDeleteCurrentBranch(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "noms-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	config := RepositoryConfig{
		FullSnapshotInterval: 10,
		Author:               "Test",
	}

	repo, err := InitRepository(tempDir, config)
	if err != nil {
		t.Fatalf("InitRepository failed: %v", err)
	}

	// Create and set current branch
	branchName := testBranchName
	err = repo.CreateBranch(branchName, "")
	if err != nil {
		t.Fatalf("CreateBranch failed: %v", err)
	}
	err = repo.SetCurrentBranch(branchName)
	if err != nil {
		t.Fatalf("SetCurrentBranch failed: %v", err)
	}

	// Try to delete current branch (should fail)
	err = repo.DeleteBranch(branchName)
	if err == nil {
		t.Error("Deleting current branch should fail")
	}
}

func TestGetCurrentBranch(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "noms-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	config := RepositoryConfig{
		FullSnapshotInterval: 10,
		Author:               "Test",
	}

	repo, err := InitRepository(tempDir, config)
	if err != nil {
		t.Fatalf("InitRepository failed: %v", err)
	}

	// Create and set a branch
	branchName := testBranchName
	err = repo.CreateBranch(branchName, "")
	if err != nil {
		t.Fatalf("CreateBranch failed: %v", err)
	}
	err = repo.SetCurrentBranch(branchName)
	if err != nil {
		t.Fatalf("SetCurrentBranch failed: %v", err)
	}

	// Get current branch
	currentBranch, err := repo.GetCurrentBranch()
	if err != nil {
		t.Fatalf("GetCurrentBranch failed: %v", err)
	}

	if currentBranch != branchName {
		t.Errorf("Expected current branch '%s', got '%s'", branchName, currentBranch)
	}
}
