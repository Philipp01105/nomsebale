package log

import (
	"fmt"
	"noms/pkg/utils"
	"noms/pkg/vcs"
	"os"
	"sort"
	"strings"
)

// Log displays the commit history
func Log() {
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

	// Check if there are any commits
	if repo.HEAD == "" {
		fmt.Println("No commits yet")
		return
	}

	// Get current branch
	currentBranch, _ := repo.GetCurrentBranch()
	if currentBranch != "" {
		fmt.Printf("Branch: %s\n\n", currentBranch)
	} else {
		fmt.Printf("HEAD detached at %s\n\n", utils.TruncateID(repo.HEAD))
	}

	// Walk through commit history
	currentID := repo.HEAD
	for currentID != "" {
		commit, err := repo.LoadCommit(currentID)
		if err != nil {
			fmt.Printf("Error loading commit %s: %v\n", utils.TruncateID(currentID), err)
			return
		}

		// Print commit information
		fmt.Printf("commit %s\n", commit.ID)
		fmt.Printf("Type: %s\n", commit.Type)
		fmt.Printf("Author: %s\n", commit.Metadata.Author)
		fmt.Printf("Date: %s\n", commit.Timestamp.Format("Mon Jan 2 15:04:05 2006"))
		fmt.Printf("Commit #%d\n", commit.CommitNumber)
		fmt.Printf("\n    %s\n\n", commit.Message)

		// Move to parent commit
		currentID = commit.ParentID
	}
}

// commitNode represents a commit in the graph
type commitNode struct {
	commit   *vcs.Commit
	branches []string // branches pointing to this commit
	isHead   bool     // is this the current HEAD
	children []string // child commit IDs
}

// LogTree displays the commit history in a tree structure showing all branches
func LogTree() {
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

	// Check if there are any commits
	if repo.HEAD == "" {
		fmt.Println("No commits yet")
		return
	}

	// Get current branch
	currentBranch, _ := repo.GetCurrentBranch()

	// Build commit graph
	commitGraph := make(map[string]*commitNode)

	// Get all branches
	branches, err := repo.ListBranches()
	if err != nil {
		fmt.Printf("Error listing branches: %v\n", err)
		return
	}

	// Collect all commits from all branches
	for _, branch := range branches {
		if branch.CommitID == "" {
			continue
		}

		// Walk through commit history for this branch
		commitID := branch.CommitID

		for commitID != "" {
			// Load or create commit node
			if _, exists := commitGraph[commitID]; !exists {
				commit, err := repo.LoadCommit(commitID)
				if err != nil {
					fmt.Printf("Error loading commit %s: %v\n", utils.TruncateID(commitID), err)
					break
				}

				node := &commitNode{
					commit:   commit,
					branches: []string{},
					children: []string{},
				}
				commitGraph[commitID] = node
			}

			// Move to parent
			commit := commitGraph[commitID].commit
			commitID = commit.ParentID
		}
	}

	// Assign branches to their HEAD commits
	for _, branch := range branches {
		if branch.CommitID != "" {
			if node, exists := commitGraph[branch.CommitID]; exists {
				node.branches = append(node.branches, branch.Name)
				if branch.Name == currentBranch {
					node.isHead = true
				}
			}
		}
	}

	// Build parent-child relationships
	for commitID, node := range commitGraph {
		if node.commit.ParentID != "" {
			if parentNode, exists := commitGraph[node.commit.ParentID]; exists {
				parentNode.children = append(parentNode.children, commitID)
			}
		}
	}

	// Collect all unique commit IDs and sort by commit number (descending)
	var allCommitIDs []string
	for commitID := range commitGraph {
		allCommitIDs = append(allCommitIDs, commitID)
	}

	sort.Slice(allCommitIDs, func(i, j int) bool {
		return commitGraph[allCommitIDs[i]].commit.CommitNumber > commitGraph[allCommitIDs[j]].commit.CommitNumber
	})

	// Display the tree
	fmt.Println("Commit history (tree view):")
	fmt.Println()

	for idx, commitID := range allCommitIDs {
		node := commitGraph[commitID]

		// Determine tree character based on position and children
		var treeChar string
		if len(node.children) > 1 {
			// This is a fork point (multiple children)
			treeChar = "*"
		} else {
			treeChar = "*"
		}

		// Build branch info string
		var branchInfo string
		if len(node.branches) > 0 {
			branchNames := make([]string, len(node.branches))
			for i, b := range node.branches {
				if b == currentBranch {
					branchNames[i] = "HEAD -> " + b
				} else {
					branchNames[i] = b
				}
			}
			// Sort branch names for consistent output
			sort.Strings(branchNames)
			branchInfo = " (" + strings.Join(branchNames, ", ") + ")"
		}

		// Print commit info
		fmt.Printf("%s commit %s%s\n", treeChar, utils.TruncateID(node.commit.ID), branchInfo)

		// Determine if we should draw a vertical line to next commit
		drawLine := idx < len(allCommitIDs)-1
		lineChar := "|"

		// For commits with multiple children, show a split
		if len(node.children) > 1 {
			lineChar = "|"
		}

		// Print additional commit details with proper indentation
		if drawLine {
			fmt.Printf("%s Author: %s\n", lineChar, node.commit.Metadata.Author)
			fmt.Printf("%s Date:   %s\n", lineChar, node.commit.Timestamp.Format("Mon Jan 2 15:04:05 2006"))
			fmt.Printf("%s\n", lineChar)
			fmt.Printf("%s     %s\n", lineChar, node.commit.Message)
		} else {
			fmt.Printf("  Author: %s\n", node.commit.Metadata.Author)
			fmt.Printf("  Date:   %s\n", node.commit.Timestamp.Format("Mon Jan 2 15:04:05 2006"))
			fmt.Printf("\n")
			fmt.Printf("      %s\n", node.commit.Message)
		}

		// Add spacing between commits
		if drawLine {
			fmt.Printf("%s\n", lineChar)
		}
	}

	fmt.Println()
}
