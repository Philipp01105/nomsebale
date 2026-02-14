package vcs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// NomsDir is the directory where VCS data is stored
	NomsDir = ".noms"
	// CommitsDir stores commit objects
	CommitsDir = "commits"
	// TreesDir stores tree state objects
	TreesDir = "trees"
	// ConfigFile stores repository configuration
	ConfigFile = "config.json"
	// HeadFile stores the current HEAD commit ID
	HeadFile = "HEAD"
)

// Repository represents a version control repository
type Repository struct {
	Path          string           `json:"path"`
	HEAD          string           `json:"head"`
	CommitCount   int              `json:"commit_count"`
	Config        RepositoryConfig `json:"config"`
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
	dirs := []string{CommitsDir, TreesDir}
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

// loadHEAD loads the HEAD reference
func (r *Repository) loadHEAD() error {
	headPath := filepath.Join(r.Path, NomsDir, HeadFile)
	data, err := os.ReadFile(headPath)
	if err != nil {
		return err
	}
	r.HEAD = string(data)
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

	// Update HEAD
	if err := r.updateHEAD(commit.ID); err != nil {
		return nil, fmt.Errorf("failed to update HEAD: %w", err)
	}

	return commit, nil
}
