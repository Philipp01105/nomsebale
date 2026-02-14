# Noms - A Simple Version Control System

Noms is a basic version control system written in Go, designed to track file changes and manage commit history.

## Features

- **Repository initialization**: Create a new version control repository
- **Commit tracking**: Create full and differential commits
- **History management**: View commit history with timestamps and metadata
- **Status checking**: See what files have changed since the last commit
- **Commit checkout**: Switch between different commits (basic implementation)

## Installation

Build the project from source:

```bash
go build -o noms cmd/main.go
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
- Repository configuration

### Check Status

View the current status of your working directory:

```bash
noms status
```

This shows:
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

### Checkout a Commit

Switch to a specific commit (basic implementation):

```bash
noms checkout <commit-id>
```

You can use either:
- Full commit ID: `081ad6fe1c66819cadd7df0843bf1bb46a0878aade3b39479469490f1a566e97`
- Partial commit ID (at least 4 characters): `081ad6fe` or `081a`

**Note**: The current implementation updates the HEAD pointer but does not restore file contents. This is a basic version control system focused on tracking changes rather than full file restoration.

## Repository Structure

```
.noms/
├── config.json       # Repository configuration and metadata
├── HEAD              # Current HEAD commit reference
├── commits/          # Commit metadata files
└── trees/            # Tree state files (file structure snapshots)
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

### Change Detection

The status command compares the current working directory against the last commit to detect:
- Modified files (changed hash)
- Added files (new paths)
- Deleted files (missing paths)

## Example Workflow

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

## Limitations

This is a basic version control system with some limitations:

1. **File restoration**: Checkout command updates HEAD but doesn't restore file contents
2. **No branching**: Single linear commit history only
3. **No merging**: Cannot merge different development paths
4. **No remote repositories**: Local-only version control
5. **No staging area**: All changes are committed together
6. **No file content storage**: Only tracks file metadata and hashes

## Future Enhancements

Potential improvements for a more complete version control system:

- Store actual file contents for full checkout capability
- Implement branching and merging
- Add remote repository support
- Implement a staging area for selective commits
- Add diff view between commits
- Support for .nomsignore files
- Optimize storage with compression
- Add tagging system for releases

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
  initializer/
    init.go            # Repository initialization
  commit/
    commit.go          # Commit creation logic
  log/
    log.go             # Log display
  status/
    status.go          # Status checking
  checkout/
    checkout.go        # Commit checkout
```

## License

This is a learning project demonstrating basic version control concepts.
