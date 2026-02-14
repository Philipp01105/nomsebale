package main

import (
	"fmt"
	"noms/pkg/commit"
	"noms/pkg/initializer"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: noms <command> [args]")
		fmt.Println("Commands:")
		fmt.Println("  init           Initialize a new noms repository")
		fmt.Println("  commit <msg>   Create a new commit with the given message")
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
	}

	if command, exists := commands[option]; exists {
		command()
	} else {
		fmt.Printf("Unknown command: %s\n", option)
		os.Exit(1)
	}
}
