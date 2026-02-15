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

// printCommitTree recursively prints commits in tree format.
// It uses depth-first traversal to display commits and their parent relationships.
// When a parent has multiple children (branch divergence), it displays them with
// proper indentation and merge point indicators.
//
// Parameters:
//   - commitGraph: map of commit IDs to commit nodes
//   - commitID: the ID of the commit to print
//   - printed: map tracking which commits have been printed to avoid duplicates
//   - currentBranch: name of the current branch, used to display HEAD indicator
func printCommitTree(commitGraph map[string]*commitNode, commitID string, printed map[string]bool, currentBranch string) {
	// Skip if already printed
	if printed[commitID] {
		return
	}
	printed[commitID] = true

	node, exists := commitGraph[commitID]
	if !exists {
		return
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
	fmt.Printf("* commit %s%s\n", utils.TruncateID(node.commit.ID), branchInfo)

	// Print commit details
	fmt.Printf("| Author: %s\n", node.commit.Metadata.Author)
	fmt.Printf("| Date:   %s\n", node.commit.Timestamp.Format("Mon Jan 2 15:04:05 2006"))
	fmt.Printf("|\n")
	fmt.Printf("|     %s\n", node.commit.Message)

	// Handle parent commit
	if node.commit.ParentID != "" {
		parentNode, parentExists := commitGraph[node.commit.ParentID]

		// Check if parent has multiple children (branch point)
		if parentExists && len(parentNode.children) > 1 {
			// This is a merge point - show the branch divergence
			fmt.Printf("|\n")

			// Find which child we are
			childIndex := -1
			for i, childID := range parentNode.children {
				if childID == commitID {
					childIndex = i
					break
				}
			}

			// If we're not the last child, show we're branching off
			if childIndex < len(parentNode.children)-1 {
				// Show other branches continuing
				for i := childIndex + 1; i < len(parentNode.children); i++ {
					otherChildID := parentNode.children[i]
					if !printed[otherChildID] {
						otherNode, otherExists := commitGraph[otherChildID]
						if !otherExists {
							continue
						}
						var otherBranchInfo string
						if len(otherNode.branches) > 0 {
							otherBranchNames := make([]string, len(otherNode.branches))
							for j, b := range otherNode.branches {
								if b == currentBranch {
									otherBranchNames[j] = "HEAD -> " + b
								} else {
									otherBranchNames[j] = b
								}
							}
							sort.Strings(otherBranchNames)
							otherBranchInfo = " (" + strings.Join(otherBranchNames, ", ") + ")"
						}
						fmt.Printf("| * commit %s%s\n", utils.TruncateID(otherNode.commit.ID), otherBranchInfo)
						fmt.Printf("| | Author: %s\n", otherNode.commit.Metadata.Author)
						fmt.Printf("| | Date:   %s\n", otherNode.commit.Timestamp.Format("Mon Jan 2 15:04:05 2006"))
						fmt.Printf("| |\n")
						fmt.Printf("| |     %s\n", otherNode.commit.Message)
						fmt.Printf("| |\n")
						printed[otherChildID] = true
					}
				}
				fmt.Printf("|/\n")
			}
		} else {
			fmt.Printf("|\n")
		}

		// Print parent commit
		printCommitTree(commitGraph, node.commit.ParentID, printed, currentBranch)
	}
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

	// Sort children by commit number (higher first) for consistent display
	for _, node := range commitGraph {
		if len(node.children) > 1 {
			sort.Slice(node.children, func(i, j int) bool {
				childI, existsI := commitGraph[node.children[i]]
				childJ, existsJ := commitGraph[node.children[j]]
				if !existsI || !existsJ {
					return existsI // Put valid children first
				}
				return childI.commit.CommitNumber > childJ.commit.CommitNumber
			})
		}
	}

	// Display the tree
	fmt.Println("Commit history (tree view):")
	fmt.Println()

	// Track which commits we've already printed
	printed := make(map[string]bool)

	// Start with commits that have no parents (root commits) or from branch heads
	var startCommits []string
	for _, branch := range branches {
		if branch.CommitID != "" && !printed[branch.CommitID] {
			startCommits = append(startCommits, branch.CommitID)
		}
	}

	// Sort start commits by commit number (descending)
	sort.Slice(startCommits, func(i, j int) bool {
		commitI, existsI := commitGraph[startCommits[i]]
		commitJ, existsJ := commitGraph[startCommits[j]]
		if !existsI || !existsJ {
			return existsI // Put valid commits first
		}
		return commitI.commit.CommitNumber > commitJ.commit.CommitNumber
	})

	// Print commits using depth-first traversal
	for _, startID := range startCommits {
		printCommitTree(commitGraph, startID, printed, currentBranch)
	}

	fmt.Println()
}
