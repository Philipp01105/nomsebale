package main

import (
	"fmt"
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
		fmt.Println("  log               Show commit history")
		fmt.Println("  status            Show working tree status")
		fmt.Println("  checkout <id>     Checkout a specific commit")
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
		"log": log.Log,
		"status": status.Status,
		"checkout": func() {
			if len(os.Args) < 3 {
				fmt.Println("Error: commit ID required")
				fmt.Println("Usage: noms checkout <commit-id>")
				os.Exit(1)
			}
			commitID := os.Args[2]
			checkout.Checkout(commitID)
		},
	}

	if command, exists := commands[option]; exists {
		command()
	} else {
		fmt.Printf("Unknown command: %s\n", option)
		os.Exit(1)
	}
}
