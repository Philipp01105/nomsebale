package vcs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// NomsDir is the directory where VCS data is stored
	NomsDir = ".noms"
	// CommitsDir stores commit objects
	CommitsDir = "commits"
	// TreesDir stores tree state objects
	TreesDir = "trees"
	// ObjectsDir stores file content blobs
	ObjectsDir = "objects"
	// ConfigFile stores repository configuration
	ConfigFile = "config.json"
	// HeadFile stores the current HEAD commit ID
	HeadFile = "HEAD"
)

// Repository represents a version control repository
type Repository struct {
	Path        string           `json:"path"`
	HEAD        string           `json:"head"`
	CommitCount int              `json:"commit_count"`
	Config      RepositoryConfig `json:"config"`
}

// RepositoryConfig contains repository configuration
type RepositoryConfig struct {
	// FullSnapshotInterval determines after how many commits a full snapshot is created
	FullSnapshotInterval int `json:"full_snapshot_interval"`
	// Author default author for commits
	Author string `json:"author"`
}

// InitRepository initializes a new repository at the given path
func InitRepository(path string, config RepositoryConfig) (*Repository, error) {
	// Create .noms directory structure
	nomsPath := filepath.Join(path, NomsDir)
	if err := os.MkdirAll(nomsPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .noms directory: %w", err)
	}

	// Create subdirectories
	dirs := []string{CommitsDir, TreesDir, ObjectsDir}
	for _, dir := range dirs {
		dirPath := filepath.Join(nomsPath, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create %s directory: %w", dir, err)
		}
	}

	// Set default config values if not provided
	if config.FullSnapshotInterval <= 0 {
		config.FullSnapshotInterval = 10 // Default: full snapshot every 10 commits
	}
	if config.Author == "" {
		config.Author = "Unknown"
	}

	repo := &Repository{
		Path:        path,
		HEAD:        "",
		CommitCount: 0,
		Config:      config,
	}

	// Save repository configuration
	if err := repo.saveConfig(); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	// Initialize HEAD file
	if err := repo.updateHEAD(""); err != nil {
		return nil, fmt.Errorf("failed to initialize HEAD: %w", err)
	}

	return repo, nil
}

// LoadRepository loads an existing repository from the given path
func LoadRepository(path string) (*Repository, error) {
	nomsPath := filepath.Join(path, NomsDir)

	// Check if .noms directory exists
	if _, err := os.Stat(nomsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("not a noms repository: %s", path)
	}

	repo := &Repository{
		Path: path,
	}

	// Load configuration
	if err := repo.loadConfig(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Load HEAD
	if err := repo.loadHEAD(); err != nil {
		return nil, fmt.Errorf("failed to load HEAD: %w", err)
	}

	return repo, nil
}

// saveConfig saves the repository configuration to disk
func (r *Repository) saveConfig() error {
	configPath := filepath.Join(r.Path, NomsDir, ConfigFile)
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

// SaveConfig is a public wrapper for saveConfig
func (r *Repository) SaveConfig() error {
	return r.saveConfig()
}

// loadConfig loads the repository configuration from disk
func (r *Repository) loadConfig() error {
	configPath := filepath.Join(r.Path, NomsDir, ConfigFile)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, r)
}

// updateHEAD updates the HEAD reference
func (r *Repository) updateHEAD(commitID string) error {
	headPath := filepath.Join(r.Path, NomsDir, HeadFile)
	r.HEAD = commitID
	return os.WriteFile(headPath, []byte(commitID), 0644)
}

// UpdateHEAD is a public wrapper for updateHEAD
func (r *Repository) UpdateHEAD(commitID string) error {
	return r.updateHEAD(commitID)
}

// loadHEAD loads the HEAD reference
func (r *Repository) loadHEAD() error {
	headPath := filepath.Join(r.Path, NomsDir, HeadFile)
	data, err := os.ReadFile(headPath)
	if err != nil {
		return err
	}

	headContent := strings.TrimSpace(string(data))

	// Check if HEAD is a symbolic reference to a branch
	if strings.HasPrefix(headContent, "ref: refs/heads/") {
		branchName := strings.TrimPrefix(headContent, "ref: refs/heads/")
		// Try to load the branch to get the actual commit ID
		branch, err := r.GetBranch(branchName)
		if err != nil {
			// Branch reference exists but branch file may not be loaded yet or has no commits
			// This can happen during initialization or if the branch file is missing
			r.HEAD = ""
			return nil
		}
		r.HEAD = branch.CommitID
	} else {
		// Direct commit reference (detached HEAD)
		r.HEAD = headContent
	}

	return nil
}

// ShouldCreateFullSnapshot determines if the next commit should be a full snapshot
func (r *Repository) ShouldCreateFullSnapshot() bool {
	// First commit is always a full snapshot
	if r.CommitCount == 0 {
		return true
	}
	// Create full snapshot at configured intervals (e.g., commit 10, 20, 30...)
	return (r.CommitCount+1)%r.Config.FullSnapshotInterval == 0
}

// SaveCommit saves a commit to disk
func (r *Repository) SaveCommit(commit *Commit) error {
	commitPath := filepath.Join(r.Path, NomsDir, CommitsDir, commit.ID+".json")
	data, err := json.MarshalIndent(commit, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal commit: %w", err)
	}
	if err := os.WriteFile(commitPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write commit file: %w", err)
	}
	return nil
}

// LoadCommit loads a commit from disk
func (r *Repository) LoadCommit(commitID string) (*Commit, error) {
	commitPath := filepath.Join(r.Path, NomsDir, CommitsDir, commitID+".json")
	data, err := os.ReadFile(commitPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read commit file: %w", err)
	}
	var commit Commit
	if err := json.Unmarshal(data, &commit); err != nil {
		return nil, fmt.Errorf("failed to unmarshal commit: %w", err)
	}
	return &commit, nil
}

// SaveTreeState saves a tree state to disk
func (r *Repository) SaveTreeState(tree *TreeState) error {
	treePath := filepath.Join(r.Path, NomsDir, TreesDir, tree.ID+".json")
	data, err := json.MarshalIndent(tree, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tree state: %w", err)
	}
	if err := os.WriteFile(treePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write tree state file: %w", err)
	}
	return nil
}

// LoadTreeState loads a tree state from disk
func (r *Repository) LoadTreeState(treeID string) (*TreeState, error) {
	treePath := filepath.Join(r.Path, NomsDir, TreesDir, treeID+".json")
	data, err := os.ReadFile(treePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tree state file: %w", err)
	}
	var tree TreeState
	if err := json.Unmarshal(data, &tree); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tree state: %w", err)
	}
	return &tree, nil
}

// CreateCommit creates a new commit in the repository
func (r *Repository) CreateCommit(treeState *TreeState, message string) (*Commit, error) {
	// Save the tree state
	if err := r.SaveTreeState(treeState); err != nil {
		return nil, fmt.Errorf("failed to save tree state: %w", err)
	}

	// Determine commit type
	var commitType CommitType
	if r.ShouldCreateFullSnapshot() {
		commitType = FullBackup
	} else {
		commitType = Differential
	}

	// Create metadata
	metadata := Metadata{
		Author: r.Config.Author,
	}

	// Create the commit
	commit := NewCommit(
		commitType,
		r.HEAD,
		treeState.ID,
		message,
		metadata,
		r.CommitCount+1,
	)

	// Save the commit
	if err := r.SaveCommit(commit); err != nil {
		return nil, fmt.Errorf("failed to save commit: %w", err)
	}

	// Update repository state
	r.CommitCount++
	r.HEAD = commit.ID

	// Save updated config
	if err := r.saveConfig(); err != nil {
		return nil, fmt.Errorf("failed to save updated config: %w", err)
	}

	// Note: HEAD file update is handled by the caller
	// (either updates branch ref or detached HEAD)

	return commit, nil
}

// SaveBlob saves file content to the objects directory
// The blob is stored with its hash as the filename
func (r *Repository) SaveBlob(hash string, content []byte) error {
	// Use first 2 characters of hash as subdirectory (sharding)
	if len(hash) < 2 {
		return fmt.Errorf("invalid hash: too short")
	}

	objDir := filepath.Join(r.Path, NomsDir, ObjectsDir, hash[:2])
	if err := os.MkdirAll(objDir, 0755); err != nil {
		return fmt.Errorf("failed to create object directory: %w", err)
	}

	objPath := filepath.Join(objDir, hash[2:])
	if err := os.WriteFile(objPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write blob: %w", err)
	}

	return nil
}

// LoadBlob loads file content from the objects directory
func (r *Repository) LoadBlob(hash string) ([]byte, error) {
	if len(hash) < 2 {
		return nil, fmt.Errorf("invalid hash: too short")
	}

	objPath := filepath.Join(r.Path, NomsDir, ObjectsDir, hash[:2], hash[2:])
	content, err := os.ReadFile(objPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read blob: %w", err)
	}

	return content, nil
}

// BlobExists checks if a blob exists in the object store
func (r *Repository) BlobExists(hash string) bool {
	if len(hash) < 2 {
		return false
	}

	objPath := filepath.Join(r.Path, NomsDir, ObjectsDir, hash[:2], hash[2:])
	_, err := os.Stat(objPath)
	return err == nil
}
