# pluginctl - Mattermost Plugin Development CLI

## Project Overview
`pluginctl` is a command-line interface tool for Mattermost plugin development. It provides utilities to manage, inspect, and work with Mattermost plugins from the command line.

## Architecture Guidelines

### Project Structure
```
pluginctl/
├── cmd/pluginctl/main.go      # CLI entrypoint with command routing
├── plugin.go                  # Plugin manifest handling utilities
├── info.go                    # Info command implementation
├── [command].go               # Additional command implementations
├── go.mod                     # Go module definition
├── go.sum                     # Go module dependencies
├── pluginctl                  # Built binary (gitignored)
└── CLAUDE.md                  # This architecture document
```

### Design Principles

#### 1. **Separation of Concerns**
- **CLI Framework**: `cmd/pluginctl/main.go` handles argument parsing, command routing, and error handling
- **Command Implementation**: Each command gets its own file (e.g., `info.go`, `build.go`, `deploy.go`) **IN THE ROOT FOLDER, NOT IN cmd/pluginctl/**
- **Utility Functions**: Common plugin operations in `plugin.go`

**CRITICAL ARCHITECTURE RULE**: ALL COMMAND LOGIC MUST BE IN SEPARATE FILES IN THE ROOT FOLDER. The cmd/pluginctl/main.go file should ONLY contain CLI framework code (argument parsing, command routing, wrapper functions). Never put command implementation logic directly in main.go.

#### 2. **Plugin Manifest Handling**
- **Always use official Mattermost types**: Import `github.com/mattermost/mattermost/server/public/model` and use `model.Manifest`
- **Validation**: Always validate plugin.json existence and format before operations
- **Path handling**: Support both current directory and custom path operations

#### 3. **Command Structure**
- **Main command router**: Add new commands to the `runCommand()` function in `cmd/pluginctl/main.go`
- **Command functions**: Name pattern: `run[Command]Command(args []string, pluginPath string) error`
- **Error handling**: Return descriptive errors, let main.go handle exit codes
- **Command implementation**: Each command's logic goes in a separate file in the root folder (e.g., `enable.go`, `disable.go`, `reset.go`)
- **Command wrapper functions**: The main.go file contains simple wrapper functions that call the actual command implementations

#### 4. **Code Organization**
- **No inline implementations**: Keep command logic in separate files
- **Reusable utilities**: Common operations go in `plugin.go`
- **Self-contained**: Each command file should be importable and testable

### Current Commands

#### `info`
- **Purpose**: Display plugin manifest information
- **Implementation**: `info.go`
- **Usage**: `pluginctl info`
- **Features**:
  - Shows plugin ID, name, version
  - Displays minimum Mattermost version
  - Indicates server/webapp code presence
  - Shows settings schema availability
- **Path Resolution**: Uses global path logic (--plugin-path flag, environment variable, or current directory)

### Adding New Commands

#### Step 1: Create Command File
Create a new file named `[command].go` with the command implementation:

```go
package main

import (
    "fmt"
    "github.com/mattermost/mattermost/server/public/model"
)

func run[Command]Command(args []string, pluginPath string) error {
    // Command implementation here
    // Use pluginPath to load plugin manifest
    return nil
}
```

#### Step 2: Register in Main Router
Add the command to the `runCommand()` function in `cmd/pluginctl/main.go`:

```go
func runCommand(command string, args []string, pluginPath string) error {
    switch command {
    case "info":
        return runInfoCommand(args, pluginPath)
    case "new-command":
        return runNewCommandCommand(args, pluginPath)
    // ... other commands
    }
}
```

#### Step 3: Update Help Text
Add the command to the `showUsage()` function in `main.go`.

### Dependencies

#### Core Dependencies
- `github.com/mattermost/mattermost/server/public/model` - Official Mattermost plugin types
- Standard Go library for CLI operations

#### Dependency Management
- Use `go mod tidy` to manage dependencies
- Prefer standard library over external packages when possible
- Only add dependencies that provide significant value

### Build System and Development Tools

#### Tool Versions
The project uses pinned versions for reproducible builds:
- **golangci-lint**: v1.62.2
- **goreleaser**: v2.6.2  
- **gosec**: v2.22.0
- **Go**: 1.24.3

#### Makefile Targets

**Development Workflow:**
- `make dev` - Quick development build (fmt, lint, build)
- `make check-changes` - Check changes (lint, security, test)
- `make verify` - Full verification (clean, lint, test, build)

**Building:**
- `make build` - Build binary for current platform
- `make build-all` - Build for all supported platforms
- `make install` - Install binary to GOPATH/bin

**Testing and Quality:**
- `make test` - Run tests
- `make test-coverage` - Run tests with coverage report
- `make lint` - Run linter
- `make lint-fix` - Fix linting issues automatically
- `make security` - Run security scan with gosec

**Development Setup:**
- `make dev-setup` - Install all development tools with pinned versions
- `make deps` - Install/update dependencies
- `make fmt` - Format code

**Release Management:**
- `make release` - Create production release (requires goreleaser)
- `make snapshot` - Create snapshot release for testing

**Utilities:**
- `make clean` - Clean build artifacts
- `make version` - Show version and tool information
- `make help` - Show all available targets

#### Configuration Files

**Makefile**
- Uses `go get -tool` for Go 1.24+ tool management
- Cross-platform build support (Linux, macOS, Windows)
- Git-based version information in binaries

**.goreleaser.yml**
- Multi-platform release automation
- GitHub releases with changelog generation
- Package manager integration (Homebrew, Scoop)
- Docker image building support

**.golangci.yml**
- 40+ enabled linters for comprehensive code quality
- Optimized for Go 1.24
- Security scanning integration
- Test file exclusions for appropriate linters

#### Development Workflow

1. **Setup**: `make dev-setup` (one-time)
2. **Development**: `make dev` (format, lint, build)
3. **Before commit**: `make check-changes` (lint, security, test)
4. **Full verification**: `make verify` (complete build verification)

#### Building
```bash
# Quick build
make build

# Cross-platform builds
make build-all

# Development build with checks
make dev
```

#### Testing
- Always test with a sample plugin.json file
- Test both current directory and custom path operations
- Verify help and version commands work correctly
- Use `make test-coverage` for coverage reports

### Error Handling Standards

#### Error Messages
- Use descriptive error messages that help users understand what went wrong
- Include file paths in error messages when relevant
- Wrap errors with context using `fmt.Errorf("operation failed: %w", err)`

#### Exit Codes
- `0`: Success
- `1`: General error
- Let main.go handle all exit codes - command functions should return errors

### Plugin Path Resolution

#### Priority Order
1. **Command-line flag**: `--plugin-path /path/to/plugin`
2. **Environment variable**: `PLUGINCTL_PLUGIN_PATH=/path/to/plugin`
3. **Current directory**: Default fallback

#### Implementation
- `getEffectivePluginPath(flagPath string) string` - Determines effective plugin path
- All commands receive the resolved plugin path as a parameter
- Path is resolved to absolute path before use

### Plugin Validation

#### Required Checks
- Plugin.json file must exist
- Plugin.json must be valid JSON
- Plugin.json must conform to Mattermost manifest schema

#### Utility Functions (plugin.go)
- `LoadPluginManifest()` - Load from current directory
- `LoadPluginManifestFromPath(path)` - Load from specific path
- `HasServerCode(manifest)` - Check for server-side code
- `HasWebappCode(manifest)` - Check for webapp code
- `IsValidPluginDirectory()` - Validate current directory

### Future Command Ideas
- `init` - Initialize a new plugin project
- `build` - Build plugin for distribution
- `deploy` - Deploy plugin to Mattermost instance
- `validate` - Validate plugin structure and manifest
- `package` - Package plugin for distribution
- `test` - Run plugin tests

### Version Management
- Current version: 0.1.0
- Update version in `main.go` when releasing
- Follow semantic versioning

### Documentation Maintenance
- **CRITICAL**: Always keep README.md up to date with any changes
- When adding new commands, update both CLAUDE.md and README.md
- When changing build processes, update both architecture docs and user docs
- When adding new dependencies or tools, document them in both files
- README.md is the user-facing documentation - it must be comprehensive and current

### Notes for Claude Sessions
- Always maintain the separation between CLI framework and command implementation
- Use the official Mattermost model types - never create custom manifest structs
- Keep command implementations in separate files for maintainability
- Always validate plugin.json before performing operations
- Test new commands with the sample plugin.json file
- Follow the established error handling patterns
- Use the build system: `make check-changes` before any commits
- Use pinned tool versions for reproducible development environments