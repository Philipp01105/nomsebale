package main

import (
	"fmt"
	"os"
)

func main() {
	// Initialize the command tree
	cmdTree := NewCommandTree()

	if len(os.Args) < 2 {
		cmdTree.PrintHelp()
		os.Exit(1)
	}

	commandName := os.Args[1]

	// Execute the command with remaining arguments
	if err := cmdTree.Execute(commandName, os.Args[2:]); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
