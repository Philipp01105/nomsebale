package branch

import (
	"fmt"
	"os"

	"github.com/Philipp01105/nomsebale/pkg/utils"
	"github.com/Philipp01105/nomsebale/pkg/vcs"
)

// List displays all branches in the repository
func List() {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return
	}

	// Load repository
	repo, err := vcs.LoadRepository(cwd)
	if err != nil {
		fmt.Printf("Error: not a noms repository. Run 'noms init' first.\n")
		return
	}

	// Get current branch
	currentBranch, err := repo.GetCurrentBranch()
	if err != nil {
		fmt.Printf("Error getting current branch: %v\n", err)
		return
	}

	// List all branches
	branches, err := repo.ListBranches()
	if err != nil {
		fmt.Printf("Error listing branches: %v\n", err)
		return
	}

	if len(branches) == 0 {
		fmt.Println("No branches found")
		return
	}

	fmt.Println("Branches:")
	for _, branch := range branches {
		marker := " "
		if branch.Name == currentBranch {
			marker = "*"
		}
		commitInfo := "no commits"
		if branch.CommitID != "" {
			commitInfo = utils.TruncateID(branch.CommitID)
		}
		fmt.Printf("%s %s -> %s\n", marker, branch.Name, commitInfo)
	}
}

// Create creates a new branch
func Create(branchName string) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return
	}

	// Load repository
	repo, err := vcs.LoadRepository(cwd)
	if err != nil {
		fmt.Printf("Error: not a noms repository. Run 'noms init' first.\n")
		return
	}

	// Create branch from current HEAD
	if err := repo.CreateBranch(branchName, ""); err != nil {
		fmt.Printf("Error creating branch: %v\n", err)
		return
	}

	commitInfo := "no commits"
	if repo.HEAD != "" {
		commitInfo = utils.TruncateID(repo.HEAD)
	}
	fmt.Printf("Created branch '%s' at %s\n", branchName, commitInfo)
}

// Delete deletes a branch
func Delete(branchName string) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return
	}

	// Load repository
	repo, err := vcs.LoadRepository(cwd)
	if err != nil {
		fmt.Printf("Error: not a noms repository. Run 'noms init' first.\n")
		return
	}

	// Delete the branch
	if err := repo.DeleteBranch(branchName); err != nil {
		fmt.Printf("Error deleting branch: %v\n", err)
		return
	}

	fmt.Printf("Deleted branch '%s'\n", branchName)
}
