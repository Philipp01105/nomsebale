package vcs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// RefsDir stores branch references
	RefsDir = "refs"
	// HeadsDir stores branch head references
	HeadsDir = "heads"
	// DefaultBranch is the name of the default branch
	DefaultBranch = "main"
)

// Branch represents a branch in the repository
type Branch struct {
	Name     string `json:"name"`
	CommitID string `json:"commit_id"`
}

// GetCurrentBranch returns the name of the current branch or empty string if in detached HEAD
func (r *Repository) GetCurrentBranch() (string, error) {
	headPath := filepath.Join(r.Path, NomsDir, HeadFile)
	data, err := os.ReadFile(headPath)
	if err != nil {
		return "", fmt.Errorf("failed to read HEAD: %w", err)
	}
	
	headContent := strings.TrimSpace(string(data))
	
	// Check if HEAD is a symbolic reference to a branch
	if strings.HasPrefix(headContent, "ref: refs/heads/") {
		branchName := strings.TrimPrefix(headContent, "ref: refs/heads/")
		return branchName, nil
	}
	
	// Detached HEAD state (direct commit reference)
	return "", nil
}

// SetCurrentBranch sets the current branch (updates HEAD to point to the branch)
func (r *Repository) SetCurrentBranch(branchName string) error {
	headPath := filepath.Join(r.Path, NomsDir, HeadFile)
	headContent := fmt.Sprintf("ref: refs/heads/%s", branchName)
	
	if err := os.WriteFile(headPath, []byte(headContent), 0644); err != nil {
		return fmt.Errorf("failed to update HEAD: %w", err)
	}
	
	return nil
}

// CreateBranch creates a new branch pointing to the specified commit
func (r *Repository) CreateBranch(branchName string, commitID string) error {
	// Validate branch name
	if branchName == "" {
		return fmt.Errorf("branch name cannot be empty")
	}
	if strings.Contains(branchName, " ") || strings.Contains(branchName, "/") {
		return fmt.Errorf("invalid branch name: %s", branchName)
	}
	
	// Check if branch already exists
	if r.BranchExists(branchName) {
		return fmt.Errorf("branch '%s' already exists", branchName)
	}
	
	// If no commit ID specified, use current HEAD
	if commitID == "" {
		// Get current branch to resolve HEAD
		currentBranch, err := r.GetCurrentBranch()
		if err != nil {
			return fmt.Errorf("failed to get current branch: %w", err)
		}
		
		if currentBranch != "" {
			// We're on a branch, get its commit
			branch, err := r.GetBranch(currentBranch)
			if err != nil {
				return fmt.Errorf("failed to get current branch commit: %w", err)
			}
			commitID = branch.CommitID
		} else {
			// Detached HEAD, use direct commit reference
			commitID = r.HEAD
		}
	}
	
	// Validate commit exists
	if commitID != "" {
		if _, err := r.LoadCommit(commitID); err != nil {
			return fmt.Errorf("commit %s does not exist: %w", commitID, err)
		}
	}
	
	// Create branch reference
	branch := &Branch{
		Name:     branchName,
		CommitID: commitID,
	}
	
	return r.saveBranch(branch)
}

// GetBranch loads a branch by name
func (r *Repository) GetBranch(branchName string) (*Branch, error) {
	branchPath := filepath.Join(r.Path, NomsDir, RefsDir, HeadsDir, branchName+".json")
	
	data, err := os.ReadFile(branchPath)
	if err != nil {
		return nil, fmt.Errorf("branch '%s' not found: %w", branchName, err)
	}
	
	var branch Branch
	if err := json.Unmarshal(data, &branch); err != nil {
		return nil, fmt.Errorf("failed to parse branch data: %w", err)
	}
	
	return &branch, nil
}

// ListBranches returns all branches in the repository
func (r *Repository) ListBranches() ([]*Branch, error) {
	branchesDir := filepath.Join(r.Path, NomsDir, RefsDir, HeadsDir)
	
	// Check if directory exists
	if _, err := os.Stat(branchesDir); os.IsNotExist(err) {
		return []*Branch{}, nil
	}
	
	entries, err := os.ReadDir(branchesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read branches directory: %w", err)
	}
	
	var branches []*Branch
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		
		branchName := strings.TrimSuffix(entry.Name(), ".json")
		branch, err := r.GetBranch(branchName)
		if err != nil {
			continue // Skip invalid branches
		}
		branches = append(branches, branch)
	}
	
	return branches, nil
}

// BranchExists checks if a branch with the given name exists
func (r *Repository) BranchExists(branchName string) bool {
	branchPath := filepath.Join(r.Path, NomsDir, RefsDir, HeadsDir, branchName+".json")
	_, err := os.Stat(branchPath)
	return err == nil
}

// DeleteBranch deletes a branch
func (r *Repository) DeleteBranch(branchName string) error {
	// Check if branch exists
	if !r.BranchExists(branchName) {
		return fmt.Errorf("branch '%s' does not exist", branchName)
	}
	
	// Prevent deleting current branch
	currentBranch, err := r.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	
	if currentBranch == branchName {
		return fmt.Errorf("cannot delete current branch '%s'", branchName)
	}
	
	// Delete branch file
	branchPath := filepath.Join(r.Path, NomsDir, RefsDir, HeadsDir, branchName+".json")
	if err := os.Remove(branchPath); err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}
	
	return nil
}

// UpdateBranchCommit updates the commit that a branch points to
func (r *Repository) UpdateBranchCommit(branchName string, commitID string) error {
	branch, err := r.GetBranch(branchName)
	if err != nil {
		return err
	}
	
	branch.CommitID = commitID
	return r.saveBranch(branch)
}

// saveBranch saves a branch to disk
func (r *Repository) saveBranch(branch *Branch) error {
	// Ensure refs/heads directory exists
	branchesDir := filepath.Join(r.Path, NomsDir, RefsDir, HeadsDir)
	if err := os.MkdirAll(branchesDir, 0755); err != nil {
		return fmt.Errorf("failed to create branches directory: %w", err)
	}
	
	branchPath := filepath.Join(branchesDir, branch.Name+".json")
	data, err := json.MarshalIndent(branch, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal branch: %w", err)
	}
	
	if err := os.WriteFile(branchPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write branch file: %w", err)
	}
	
	return nil
}

// SwitchBranch switches to a different branch
func (r *Repository) SwitchBranch(branchName string) error {
	// Check if branch exists
	if !r.BranchExists(branchName) {
		return fmt.Errorf("branch '%s' does not exist", branchName)
	}
	
	// Load the branch
	branch, err := r.GetBranch(branchName)
	if err != nil {
		return err
	}
	
	// Update HEAD to point to the branch
	if err := r.SetCurrentBranch(branchName); err != nil {
		return err
	}
	
	// Update repository's HEAD to the branch's commit
	r.HEAD = branch.CommitID
	
	return nil
}
