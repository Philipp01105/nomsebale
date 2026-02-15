package log

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Philipp01105/nomsebale/pkg/utils"
	"github.com/Philipp01105/nomsebale/pkg/vcs"
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
	currentBranch, err := repo.GetCurrentBranch()
	if err != nil {
		fmt.Printf("Error getting current branch: %v\n", err)
		return
	}
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
	children []string // child commit IDs
	isHead   bool     // is this the current HEAD
}

// buildBranchInfo creates a formatted string with branch names
func buildBranchInfo(branches []string, currentBranch string) string {
	if len(branches) == 0 {
		return ""
	}
	branchNames := make([]string, len(branches))
	for i, b := range branches {
		if b == currentBranch {
			branchNames[i] = "HEAD -> " + b
		} else {
			branchNames[i] = b
		}
	}
	sort.Strings(branchNames)
	return " (" + strings.Join(branchNames, ", ") + ")"
}

// printCommitInfo prints the basic commit information
func printCommitInfo(node *commitNode, currentBranch string) {
	branchInfo := buildBranchInfo(node.branches, currentBranch)
	fmt.Printf("* commit %s%s\n", utils.TruncateID(node.commit.ID), branchInfo)
	fmt.Printf("| Author: %s\n", node.commit.Metadata.Author)
	fmt.Printf("| Date:   %s\n", node.commit.Timestamp.Format("Mon Jan 2 15:04:05 2006"))
	fmt.Printf("|\n")
	fmt.Printf("|     %s\n", node.commit.Message)
}

// printBranchDivergence prints sibling branches when there's a branch point
func printBranchDivergence(commitGraph map[string]*commitNode, parentNode *commitNode, commitID string, printed map[string]bool, currentBranch string) {
	// Find which child we are
	childIndex := -1
	for i, childID := range parentNode.children {
		if childID == commitID {
			childIndex = i
			break
		}
	}

	// If we're not the last child, show other branches
	if childIndex < len(parentNode.children)-1 {
		for i := childIndex + 1; i < len(parentNode.children); i++ {
			otherChildID := parentNode.children[i]
			if !printed[otherChildID] {
				otherNode, otherExists := commitGraph[otherChildID]
				if !otherExists {
					continue
				}
				otherBranchInfo := buildBranchInfo(otherNode.branches, currentBranch)
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
}

// printCommitTree recursively prints commits in tree format.
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

	// Print commit info
	printCommitInfo(node, currentBranch)

	// Handle parent commit
	if node.commit.ParentID != "" {
		parentNode, parentExists := commitGraph[node.commit.ParentID]

		// Check if parent has multiple children (branch point)
		if parentExists && len(parentNode.children) > 1 {
			fmt.Printf("|\n")
			printBranchDivergence(commitGraph, parentNode, commitID, printed, currentBranch)
		} else {
			fmt.Printf("|\n")
		}

		// Print parent commit
		printCommitTree(commitGraph, node.commit.ParentID, printed, currentBranch)
	}
}

// buildCommitGraph creates a graph of all commits from all branches
func buildCommitGraph(repo *vcs.Repository, branches []*vcs.Branch) (map[string]*commitNode, error) {
	commitGraph := make(map[string]*commitNode)

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
					return nil, fmt.Errorf("loading commit %s: %w", utils.TruncateID(commitID), err)
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

	return commitGraph, nil
}

// assignBranchesToCommits assigns branch names to their HEAD commits
func assignBranchesToCommits(commitGraph map[string]*commitNode, branches []*vcs.Branch, currentBranch string) {
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
}

// buildParentChildRelationships establishes parent-child links in the commit graph
func buildParentChildRelationships(commitGraph map[string]*commitNode) {
	for commitID, node := range commitGraph {
		if node.commit.ParentID != "" {
			if parentNode, exists := commitGraph[node.commit.ParentID]; exists {
				parentNode.children = append(parentNode.children, commitID)
			}
		}
	}
}

// sortChildrenByCommitNumber sorts children by commit number for consistent display
func sortChildrenByCommitNumber(commitGraph map[string]*commitNode) {
	for _, node := range commitGraph {
		if len(node.children) > 1 {
			sort.Slice(node.children, func(i, j int) bool {
				childI, existsI := commitGraph[node.children[i]]
				childJ, existsJ := commitGraph[node.children[j]]
				if !existsI || !existsJ {
					return existsI
				}
				return childI.commit.CommitNumber > childJ.commit.CommitNumber
			})
		}
	}
}

// getStartCommits returns the starting points for tree traversal
func getStartCommits(commitGraph map[string]*commitNode, branches []*vcs.Branch) []string {
	var startCommits []string
	printed := make(map[string]bool)

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
			return existsI
		}
		return commitI.commit.CommitNumber > commitJ.commit.CommitNumber
	})

	return startCommits
}

// HistoryTree displays the commit history in a tree structure showing all branches
func HistoryTree() {
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
	currentBranch, err := repo.GetCurrentBranch()
	if err != nil {
		fmt.Printf("Error getting current branch: %v\n", err)
		return
	}

	// Get all branches
	branches, err := repo.ListBranches()
	if err != nil {
		fmt.Printf("Error listing branches: %v\n", err)
		return
	}

	// Build commit graph
	commitGraph, err := buildCommitGraph(repo, branches)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Assign branches to their HEAD commits
	assignBranchesToCommits(commitGraph, branches, currentBranch)

	// Build parent-child relationships
	buildParentChildRelationships(commitGraph)

	// Sort children by commit number for consistent display
	sortChildrenByCommitNumber(commitGraph)

	// Display the tree
	fmt.Println("Commit history (tree view):")
	fmt.Println()

	// Get starting commits and print tree
	startCommits := getStartCommits(commitGraph, branches)
	printed := make(map[string]bool)

	for _, startID := range startCommits {
		printCommitTree(commitGraph, startID, printed, currentBranch)
	}

	fmt.Println()
}
