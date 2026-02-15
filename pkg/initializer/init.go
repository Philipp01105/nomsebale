package initializer

import (
	"fmt"
	"noms/pkg/vcs"
	"os"
	"time"
)

type File struct {
	Name      string
	extension string
	hash      string
	delta     time.Duration
	changes   []change
}

type change struct {
	before string
	after  string
}

func Init() {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return
	}

	// Check if already initialized
	if _, err := vcs.LoadRepository(cwd); err == nil {
		fmt.Println("Repository already initialized in this directory")
		return
	}

	// Initialize repository with default configuration
	config := vcs.RepositoryConfig{
		FullSnapshotInterval: 10,
		Author:               "Unknown User",
	}

	repo, err := vcs.InitRepository(cwd, config)
	if err != nil {
		fmt.Printf("Error initializing repository: %v\n", err)
		return
	}

	// Create default branch (main)
	if err := repo.CreateBranch(vcs.DefaultBranch, ""); err != nil {
		fmt.Printf("Error creating default branch: %v\n", err)
		return
	}

	// Set current branch to main
	if err := repo.SetCurrentBranch(vcs.DefaultBranch); err != nil {
		fmt.Printf("Error setting current branch: %v\n", err)
		return
	}

	fmt.Println("Initialized empty noms repository in", cwd)
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Full snapshot interval: every %d commits\n", repo.Config.FullSnapshotInterval)
	fmt.Printf("  Default author: %s\n", repo.Config.Author)
	fmt.Printf("  Default branch: %s\n", vcs.DefaultBranch)
}
