package main

import (
	"fmt"
	"noms/pkg/initializer"
	"os"
)

func main() {
	var option = os.Args[1]

	var commands = map[string]func(){
		"init": initializer.Init,
		"goodbye": func() {
			fmt.Println("Goodbye, World!")
		},
	}

	if command, exists := commands[option]; exists {
		command()
	}
}
