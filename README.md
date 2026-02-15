# Noms - A Simple Version Control System

[![CI](https://github.com/Philipp01105/nomsebale/actions/workflows/ci.yml/badge.svg)](https://github.com/Philipp01105/nomsebale/actions/workflows/ci.yml)
[![Code Quality](https://github.com/Philipp01105/nomsebale/actions/workflows/code-quality.yml/badge.svg)](https://github.com/Philipp01105/nomsebale/actions/workflows/code-quality.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Philipp01105/nomsebale)](https://goreportcard.com/report/github.com/Philipp01105/nomsebale)

Noms is a basic version control system written in Go, designed to track file changes and manage commit history with support for branching.

## Features

- **Repository initialization**: Create a new version control repository
- **Branching**: Create, switch between, and delete branches for parallel development
- **File storage**: Store complete file contents as blobs
- **Commit tracking**: Create full and differential commits
- **History management**: View commit history with timestamps and metadata
- **Status checking**: See what files have changed since the last commit
- **Complete checkout**: Restore file contents from any commit or branch in history

## Installation

Build the project from source:

```bash
go build -o noms cmd/main.go
```

Or use the Makefile:

```bash
make build
```

Or install directly:

```bash
go install github.com/Philipp01105/nomsebale/cmd@latest
```

## Usage

### Initialize a Repository

Initialize a new noms repository in the current directory:

```bash
noms init
```

This creates a `.noms` directory that stores all version control data including:
- Commits metadata
- Tree states (snapshots of file structures)
- Branch references
- Repository configuration

A default `main` branch is automatically created.

### Check Status

View the current status of your working directory:

```bash
noms status
```

This shows:
- Current branch or detached HEAD state
- Current commit information
- Modified files
- New files
- Deleted files

### Create a Commit

Commit all changes in the current directory:

```bash
noms commit "Your commit message"
```

Noms automatically creates:
- **Full snapshots** for the first commit and at regular intervals (every 10 commits by default)
- **Differential commits** for incremental changes between full snapshots

### View Commit History

Display the complete commit history:

```bash
noms log
```

Shows for each commit:
- Full commit ID
- Commit type (full or differential)
- Author
- Timestamp
- Commit number
- Commit message

### Working with Branches

#### List Branches

View all branches in the repository:

```bash
noms branch
```

The current branch is marked with an asterisk (*).

#### Create a New Branch

Create a new branch from the current commit:

```bash
noms branch <branch-name>
```

Example:
```bash
noms branch feature-x
```

This creates a new branch pointing to the current HEAD commit.

#### Switch Branches

Switch to a different branch:

```bash
noms checkout <branch-name>
```

Example:
```bash
noms checkout feature-x
```

When you switch branches:
- All files are restored to the state of that branch's latest commit
- Future commits will be made on the new branch
- Each branch maintains its own independent commit history

#### Delete a Branch

Delete a branch that is no longer needed:

```bash
noms branch -d <branch-name>
```

Example:
```bash
noms branch -d feature-x
```

Note: You cannot delete the branch you're currently on.

### Checkout a Commit or Branch

Switch to a specific commit or branch and restore all files:

```bash
noms checkout <commit-id|branch-name>
```

You can use either:
- Branch name: `main`, `feature-x`
- Full commit ID: `081ad6fe1c66819cadd7df0843bf1bb46a0878aade3b39479469490f1a566e97`
- Partial commit ID (at least 4 characters): `081ad6fe` or `081a`

The checkout command:
- Restores all files to their state in the specified commit or branch
- Deletes files that don't exist in that commit
- Updates file permissions
- Updates the HEAD pointer to the specified commit or branch
- When checking out a commit directly, you enter "detached HEAD" state

## Repository Structure

```
.noms/
├── config.json       # Repository configuration and metadata
├── HEAD              # Current HEAD reference (branch or commit)
├── refs/
│   └── heads/        # Branch references
├── commits/          # Commit metadata files
├── trees/            # Tree state files (file structure snapshots)
└── objects/          # File content blobs (stored by hash)
```

## Configuration

Default configuration:
- **Full snapshot interval**: Every 10 commits
- **Default author**: Unknown User

Configuration is stored in `.noms/config.json`.

## How It Works

### Commits

Noms uses two types of commits:

1. **Full Backup Commits**: Complete snapshots of the entire file tree
   - Created for the first commit
   - Created at regular intervals (configurable, default: every 10 commits)

2. **Differential Commits**: Store only changes since the previous commit
   - More space-efficient
   - Reference parent commits to build complete file state

### Tree States

Tree states capture the structure and metadata of files at a specific point in time:
- File paths
- File hashes (SHA256)
- File sizes
- Permissions
- Modification timestamps

### File Storage

File contents are stored as blobs in the objects directory:
- Blobs are named by their SHA256 hash
- Files are sharded into subdirectories using the first 2 characters of the hash
- Duplicate files (same hash) are stored only once, saving space

### Branches

Branches allow you to maintain independent lines of development:

1. **Branch Storage**: Each branch is stored as a reference in `.noms/refs/heads/`
2. **HEAD Tracking**: The HEAD file contains either:
   - A symbolic reference to a branch: `ref: refs/heads/main`
   - A direct commit reference (detached HEAD state)
3. **Branch History**: Each branch tracks its own commit history independently
4. **Tree Splitting**: Creating a branch from a commit creates a new development path

When you commit on a branch:
- The branch reference is updated to point to the new commit
- The commit's parent is the previous commit on that branch
- Other branches remain unchanged

### Change Detection

The status command compares the current working directory against the last commit to detect:
- Modified files (changed hash)
- Added files (new paths)
- Deleted files (missing paths)

## Example Workflows

### Basic Workflow

```bash
# Initialize repository
noms init

# Create some files
echo "Hello" > hello.txt
echo "World" > world.txt

# Check status
noms status

# Create first commit
noms commit "Initial commit"

# Modify a file
echo "More content" >> hello.txt

# Check what changed
noms status

# Commit the change
noms commit "Updated hello.txt"

# View history
noms log

# Checkout previous commit
noms checkout 081ad6fe
```

### Branching Workflow

```bash
# Initialize repository and create first commit
noms init
echo "Initial content" > README.md
noms commit "Initial commit"

# Create a feature branch
noms branch feature-x

# Switch to the feature branch
noms checkout feature-x

# Make changes on the feature branch
echo "Feature X code" > feature-x.txt
noms commit "Add feature X"

# Switch back to main branch
noms checkout main

# Make changes on main branch
echo "Main branch update" > update.txt
noms commit "Main branch work"

# View all branches and their commits
noms branch

# View feature branch history
noms checkout feature-x
noms log

# Delete a branch when done
noms checkout main
noms branch -d feature-x
```

## Limitations

This is a basic version control system with some limitations:

1. **No merging**: Cannot merge different development paths (branches remain independent)
2. **No remote repositories**: Local-only version control
3. **No staging area**: All changes are committed together
4. **No incremental checkout**: Always restores the complete tree state
5. **No conflict resolution**: Manual file management required when switching branches

## Future Enhancements

Potential improvements for a more complete version control system:

- Implement branch merging with conflict resolution
- Add remote repository support
- Implement a staging area for selective commits
- Add diff view between commits and branches
- Support for .nomsignore files
- Optimize storage with compression
- Add tagging system for releases
- Implement file locking for concurrent access

## Development

The project is structured as follows:

```
cmd/
  main.go              # CLI entry point
pkg/
  vcs/
    repository.go      # Repository management
    commit.go          # Commit structures and logic
    tree.go            # Tree state and delta computation
    branch.go          # Branch management
  initializer/
    init.go            # Repository initialization
  commit/
    commit.go          # Commit creation logic
  branch/
    branch.go          # Branch operations
  log/
    log.go             # Log display
  status/
    status.go          # Status checking
  checkout/
    checkout.go        # Commit checkout
```

## Contributing

Contributions are welcome! This project uses automated workflows for code quality and testing.

### Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/Philipp01105/nomsebale.git
   cd nomsebale
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Run tests:
   ```bash
   make test
   # or
   go test ./...
   ```

### Code Quality

Before submitting a pull request, ensure your code passes all checks:

```bash
# Format code
make fmt

# Run linter
make vet

# Run tests
make test

# Run all CI checks
make ci
```

### Testing

The project includes unit tests for core functionality:

```bash
# Run all tests
go test ./...

# Run tests with coverage
make test-coverage

# Run tests with verbose output
go test -v ./...
```

### GitHub Workflows

This project uses GitHub Actions for continuous integration:

- **CI Workflow**: Runs on every push and pull request
  - Builds the project on multiple Go versions (1.21, 1.22, 1.23)
  - Runs all tests with race detection
  - Generates code coverage reports

- **Code Quality Workflow**: Ensures code quality standards
  - Checks code formatting with `gofmt`
  - Runs `go vet` for static analysis
  - Runs `golangci-lint` for comprehensive linting
  - Verifies `go mod tidy` is up to date

- **Release Workflow**: Automated releases on tags
  - Builds binaries for multiple platforms (Linux, macOS, Windows)
  - Creates GitHub releases with binaries and checksums
  - Triggers on version tags (e.g., `v1.0.0`)

## License

This is a learning project demonstrating basic version control concepts.
