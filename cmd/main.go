package main

import (
	"fmt"
	"noms/pkg/branch"
	"noms/pkg/checkout"
	"noms/pkg/commit"
	"noms/pkg/initializer"
	"noms/pkg/log"
	"noms/pkg/status"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: noms <command> [args]")
		fmt.Println("Commands:")
		fmt.Println("  init              Initialize a new noms repository")
		fmt.Println("  commit <msg>      Create a new commit with the given message")
		fmt.Println("  log [--tree]      Show commit history (--tree shows all branches in tree structure)")
		fmt.Println("  status            Show working tree status")
		fmt.Println("  checkout <ref>    Checkout a specific commit or branch")
		fmt.Println("  branch [name]     List branches or create a new branch")
		fmt.Println("  branch -d <name>  Delete a branch")
		os.Exit(1)
	}

	var option = os.Args[1]

	var commands = map[string]func(){
		"init": initializer.Init,
		"commit": func() {
			if len(os.Args) < 3 {
				fmt.Println("Error: commit message required")
				fmt.Println("Usage: noms commit <message>")
				os.Exit(1)
			}
			message := os.Args[2]
			commit.Commit(message)
		},
		"log": func() {
			// Check for --tree flag
			showTree := false
			if len(os.Args) > 2 && os.Args[2] == "--tree" {
				showTree = true
			}
			if showTree {
				log.LogTree()
			} else {
				log.Log()
			}
		},
		"status": status.Status,
		"checkout": func() {
			if len(os.Args) < 3 {
				fmt.Println("Error: commit ID or branch name required")
				fmt.Println("Usage: noms checkout <commit-id|branch-name>")
				os.Exit(1)
			}
			ref := os.Args[2]
			checkout.Checkout(ref)
		},
		"branch": func() {
			if len(os.Args) < 3 {
				// No arguments - list branches
				branch.List()
				return
			}

			// Check for -d flag (delete)
			if os.Args[2] == "-d" {
				if len(os.Args) < 4 {
					fmt.Println("Error: branch name required")
					fmt.Println("Usage: noms branch -d <branch-name>")
					os.Exit(1)
				}
				branchName := os.Args[3]
				branch.Delete(branchName)
				return
			}

			// Create new branch
			branchName := os.Args[2]
			branch.Create(branchName)
		},
	}

	if command, exists := commands[option]; exists {
		command()
	} else {
		fmt.Printf("Unknown command: %s\n", option)
		os.Exit(1)
	}
}
