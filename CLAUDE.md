# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with
code in this repository.

## Project Overview

`pass-env` is a CLI tool that integrates with `pass` (the Unix password
manager) to set environment variables from password store entries before
running commands. It maintains its own internal store and dependency tracking
system.

## Architecture

### Core Components

- **`state/`**: Manages pass-env's internal state and store
  - `index` variable (type `passNameDependents`): Maps password names from the
    main password store to their dependent cache entries
  - Persisted to `~/.local/share/pass-env/store.index` via gob encoding

- **`cmd/`**: Cobra-based CLI commands
  - `root.go`: Base command and execution
  - `init.go`: Initializes pass-env's internal store, links to existing pass
    store, and sets up the index
  - `clear.go`: Cache clearing functionality (stub)

- **`lib/`**: Utility libraries
  - `set/`: Generic set implementation using `map[any]struct{}`
  - `fs/`: Filesystem utilities for checking file/dir/symlink existence

### State Management

The `state` package maintains three key paths:
- `Store()`: `~/.local/share/pass-env/store` - pass-env's internal password store
- `PassStore()`: `~/.local/share/pass-env/.pass-store` - symlink to user's main
  pass store
- `StoreIndex()`: `~/.local/share/pass-env/store.index` - dependency tracking file

The index is loaded on package initialization if pass-env is already initialized.

## Development

### Build
```bash
go build
```

### Dependencies
The project uses Nix flakes for development environment:
```bash
nix develop
```

This provides:
- Go toolchain
- `cobra-cli` for generating new commands

### Adding New Commands
Use cobra-cli to generate command scaffolding:
```bash
cobra-cli add <command-name>
```

## Key Concepts

**Index**: The index tracks dependencies between password names and cache entries.

**Pass Integration**: This tool wraps the standard Unix `pass` command and
maintains its own separate store while symlinking to the user's main password
store for reference.
