package vcs

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// CommitType represents the type of commit
type CommitType string

const (
	// FullBackup represents a full snapshot commit
	FullBackup CommitType = "full"
	// Differential represents an incremental commit
	Differential CommitType = "differential"
)

// Commit represents a version control commit
type Commit struct {
	Timestamp    time.Time  `json:"timestamp"`
	Metadata     Metadata   `json:"metadata"`
	ID           string     `json:"id"`
	ParentID     string     `json:"parent_id,omitempty"`
	TreeStateID  string     `json:"tree_state_id"`
	Message      string     `json:"message"`
	Type         CommitType `json:"type"`
	CommitNumber int        `json:"commit_number"`
}

// Metadata contains additional commit information
type Metadata struct {
	Extra  map[string]string `json:"extra,omitempty"`
	Author string            `json:"author"`
}

// NewCommit creates a new commit with the given parameters
func NewCommit(commitType CommitType, parentID, treeStateID, message string, metadata Metadata, commitNumber int) *Commit {
	commit := &Commit{
		Type:         commitType,
		Timestamp:    time.Now(),
		ParentID:     parentID,
		TreeStateID:  treeStateID,
		Message:      message,
		Metadata:     metadata,
		CommitNumber: commitNumber,
	}
	commit.ID = commit.generateID()
	return commit
}

// generateID generates a unique hash for the commit based on metadata
// Hash = SHA256(parent_id + tree_state_id + timestamp + type + message)
func (c *Commit) generateID() string {
	data := fmt.Sprintf("%s%s%d%s%s",
		c.ParentID,
		c.TreeStateID,
		c.Timestamp.Unix(),
		c.Type,
		c.Message,
	)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// IsFullBackup returns true if this is a full backup commit
func (c *Commit) IsFullBackup() bool {
	return c.Type == FullBackup
}

// IsDifferential returns true if this is a differential commit
func (c *Commit) IsDifferential() bool {
	return c.Type == Differential
}
