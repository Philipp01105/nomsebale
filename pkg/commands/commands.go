package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/Philipp01105/nomsebale/pkg/branch"
	"github.com/Philipp01105/nomsebale/pkg/checkout"
	"github.com/Philipp01105/nomsebale/pkg/commit"
	"github.com/Philipp01105/nomsebale/pkg/initializer"
	"github.com/Philipp01105/nomsebale/pkg/log"
	"github.com/Philipp01105/nomsebale/pkg/status"
)

// GlobalFlags defines flags that can be used with any command
type GlobalFlags struct {
	Verbose bool
	Help    bool
}

// CommandDefinition defines a command with its flags and handler
type CommandDefinition struct {
	Name        string
	Description string
	Usage       string
	Flags       map[string]FlagDefinition
	Handler     func(*flag.FlagSet) error
}

// FlagDefinition defines a single flag
type FlagDefinition struct {
	Name        string
	ShortName   string
	Description string
	Type        string // "bool", "string", "int"
	Default     interface{}
}

// CommandTree holds all command definitions
type CommandTree struct {
	GlobalFlags map[string]FlagDefinition
	Commands    map[string]CommandDefinition
}

// NewCommandTree creates and initializes the command tree
func NewCommandTree() *CommandTree {
	tree := &CommandTree{
		GlobalFlags: make(map[string]FlagDefinition),
		Commands:    make(map[string]CommandDefinition),
	}

	// Define global flags (available to all commands)
	tree.GlobalFlags["verbose"] = FlagDefinition{
		Name:        "verbose",
		ShortName:   "v",
		Description: "enable verbose output",
		Type:        "bool",
		Default:     false,
	}

	// Define init command
	tree.Commands["init"] = CommandDefinition{
		Name:        "init",
		Description: "Initialize a new noms repository",
		Usage:       "noms init",
		Flags:       make(map[string]FlagDefinition),
		Handler: func(fs *flag.FlagSet) error {
			initializer.Init()
			return nil
		},
	}

	// Define commit command
	tree.Commands["commit"] = CommandDefinition{
		Name:        "commit",
		Description: "Create a new commit with the given message",
		Usage:       "noms commit <message>",
		Flags:       make(map[string]FlagDefinition),
		Handler: func(fs *flag.FlagSet) error {
			args := fs.Args()
			if len(args) < 1 {
				fmt.Println("Error: commit message required")
				fmt.Println("Usage: noms commit <message>")
				os.Exit(1)
			}
			message := args[0]
			commit.Commit(message)
			return nil
		},
	}

	// Define log command with its specific flags
	tree.Commands["log"] = CommandDefinition{
		Name:        "log",
		Description: "Show commit history",
		Usage:       "noms log [-t|--tree]",
		Flags: map[string]FlagDefinition{
			"tree": {
				Name:        "tree",
				ShortName:   "t",
				Description: "show all branches in tree structure",
				Type:        "bool",
				Default:     false,
			},
		},
		Handler: func(fs *flag.FlagSet) error {
			// Flags are already parsed by the framework
			showTree := fs.Lookup("tree").Value.(flag.Getter).Get().(bool) ||
				fs.Lookup("t").Value.(flag.Getter).Get().(bool)

			if showTree {
				log.HistoryTree()
			} else {
				log.Log()
			}
			return nil
		},
	}

	// Define status command
	tree.Commands["status"] = CommandDefinition{
		Name:        "status",
		Description: "Show working tree status",
		Usage:       "noms status",
		Flags:       make(map[string]FlagDefinition),
		Handler: func(fs *flag.FlagSet) error {
			status.Status()
			return nil
		},
	}

	// Define checkout command
	tree.Commands["checkout"] = CommandDefinition{
		Name:        "checkout",
		Description: "Checkout a specific commit or branch",
		Usage:       "noms checkout <ref>",
		Flags:       make(map[string]FlagDefinition),
		Handler: func(fs *flag.FlagSet) error {
			args := fs.Args()
			if len(args) < 1 {
				fmt.Println("Error: commit ID or branch name required")
				fmt.Println("Usage: noms checkout <commit-id|branch-name>")
				os.Exit(1)
			}
			ref := args[0]
			checkout.Checkout(ref)
			return nil
		},
	}

	// Define branch command with its specific flags
	tree.Commands["branch"] = CommandDefinition{
		Name:        "branch",
		Description: "List branches or create/delete a branch",
		Usage:       "noms branch [name] [-d <name>]",
		Flags: map[string]FlagDefinition{
			"delete": {
				Name:        "delete",
				ShortName:   "d",
				Description: "delete a branch",
				Type:        "string",
				Default:     "",
			},
		},
		Handler: func(fs *flag.FlagSet) error {
			deleteBranch := fs.Lookup("delete").Value.String()
			deleteBranchShort := fs.Lookup("d").Value.String()
			args := fs.Args()

			// Check for delete flag (long or short form)
			if deleteBranch != "" || deleteBranchShort != "" {
				branchToDelete := deleteBranch
				if branchToDelete == "" {
					branchToDelete = deleteBranchShort
				}
				branch.Delete(branchToDelete)
				return nil
			}

			// Check for branch name argument (create)
			if len(args) > 0 {
				branchName := args[0]
				branch.Create(branchName)
				return nil
			}

			// No arguments - list branches
			branch.List()
			return nil
		},
	}

	return tree
}

// Execute runs the specified command with its flags
func (ct *CommandTree) Execute(commandName string, args []string) error {
	cmd, exists := ct.Commands[commandName]
	if !exists {
		return fmt.Errorf("unknown command: %s", commandName)
	}

	// Create a FlagSet for this command
	fs := flag.NewFlagSet(cmd.Name, flag.ExitOnError)

	// Add command-specific flags
	for _, flagDef := range cmd.Flags {
		switch flagDef.Type {
		case "bool":
			fs.Bool(flagDef.Name, flagDef.Default.(bool), flagDef.Description)
			if flagDef.ShortName != "" {
				fs.Bool(flagDef.ShortName, flagDef.Default.(bool), flagDef.Description)
			}
		case "string":
			fs.String(flagDef.Name, flagDef.Default.(string), flagDef.Description)
			if flagDef.ShortName != "" {
				fs.String(flagDef.ShortName, flagDef.Default.(string), flagDef.Description)
			}
		case "int":
			fs.Int(flagDef.Name, flagDef.Default.(int), flagDef.Description)
			if flagDef.ShortName != "" {
				fs.Int(flagDef.ShortName, flagDef.Default.(int), flagDef.Description)
			}
		}
	}

	// Add global flags
	for _, flagDef := range ct.GlobalFlags {
		if flagDef.Type == "bool" {
			fs.Bool(flagDef.Name, flagDef.Default.(bool), flagDef.Description)
			if flagDef.ShortName != "" {
				fs.Bool(flagDef.ShortName, flagDef.Default.(bool), flagDef.Description)
			}
		}
	}

	// Parse the flags
	err := fs.Parse(args)
	if err != nil {
		return err
	}

	// Execute the command handler
	return cmd.Handler(fs)
}

// PrintHelp prints the help message
func (ct *CommandTree) PrintHelp() {
	fmt.Println("Usage: noms <command> [args]")
	fmt.Println("\nCommands:")

	// Print commands in a consistent order
	commandOrder := []string{"init", "commit", "log", "status", "checkout", "branch"}
	for _, cmdName := range commandOrder {
		if cmd, exists := ct.Commands[cmdName]; exists {
			fmt.Printf("  %-20s %s\n", cmd.Usage, cmd.Description)

			// Print command-specific flags
			if len(cmd.Flags) > 0 {
				for _, flagDef := range cmd.Flags {
					flagStr := fmt.Sprintf("-%s, --%s", flagDef.ShortName, flagDef.Name)
					if flagDef.ShortName == "" {
						flagStr = fmt.Sprintf("--%s", flagDef.Name)
					}
					fmt.Printf("    %-18s %s\n", flagStr, flagDef.Description)
				}
			}
		}
	}

	// Print global flags
	if len(ct.GlobalFlags) > 0 {
		fmt.Println("\nGlobal Flags (available to all commands):")
		for _, flagDef := range ct.GlobalFlags {
			flagStr := fmt.Sprintf("-%s, --%s", flagDef.ShortName, flagDef.Name)
			fmt.Printf("  %-20s %s\n", flagStr, flagDef.Description)
		}
	}
}
