# pluginctl

A command-line interface tool for Mattermost plugin development. `pluginctl` provides utilities to manage, inspect, and work with Mattermost plugins from the command line.

## Installation

### From Source

```bash
git clone https://github.com/mattermost/pluginctl.git
cd pluginctl
make build
```

### Binary Installation

Download the latest binary from the [releases page](https://github.com/mattermost/pluginctl/releases).

### Package Managers

```bash
# Using go install
go install github.com/mattermost/pluginctl/cmd/pluginctl@latest

# Using Homebrew (if available)
brew install mattermost/tap/pluginctl

# Using Scoop on Windows (if available)
scoop bucket add mattermost https://github.com/mattermost/scoop-bucket.git
scoop install pluginctl
```

## Usage

### Basic Commands

```bash
# Display plugin information
pluginctl info

# Show help
pluginctl --help

# Show version
pluginctl --version
```

### Plugin Path Configuration

`pluginctl` supports multiple ways to specify the plugin directory:

1. **Command-line flag** (highest priority):

   ```bash
   pluginctl --plugin-path /path/to/plugin info
   ```

2. **Environment variable**:

   ```bash
   export PLUGINCTL_PLUGIN_PATH=/path/to/plugin
   pluginctl info
   ```

3. **Current directory** (default):
   ```bash
   cd /path/to/plugin
   pluginctl info
   ```

### Commands

#### `info`

Display comprehensive information about a Mattermost plugin:

```bash
pluginctl info
```

**Output includes:**

- Plugin ID, name, and version
- Minimum Mattermost server version required
- Description (if available)
- Server code presence and supported platforms
- Webapp code presence and bundle path
- Settings schema availability

**Example output:**

```
Plugin Information:
==================

ID:              com.example.testplugin
Name:            Test Plugin
Version:         1.0.0
Min MM Version:  7.0.0
Description:     A test plugin for demonstrating pluginctl functionality

Code Components:
================
Server Code:     Yes
  Executables:   linux-amd64, darwin-amd64, windows-amd64
Webapp Code:     Yes
  Bundle Path:   webapp/dist/main.js
Settings Schema: Yes
```

## Requirements

- Go 1.24.3 or later
- Valid Mattermost plugin directory with `plugin.json` manifest file

## Development Tools

The project uses the following tools for development and release automation:

- **golangci-lint** v1.62.2 - Code linting and quality checks
- **goreleaser** v2.6.2 - Automated releases and cross-platform builds
- **gosec** v2.22.0 - Security vulnerability scanning

## Plugin Directory Structure

`pluginctl` expects to work with standard Mattermost plugin directories containing a `plugin.json` file. For more information about Mattermost plugin structure, visit the [official documentation](https://developers.mattermost.com/integrate/plugins/).

## Environment Variables

| Variable                | Description                   |
| ----------------------- | ----------------------------- |
| `PLUGINCTL_PLUGIN_PATH` | Default plugin directory path |

## Contributing

We welcome contributions to `pluginctl`! Please see the [CLAUDE.md](CLAUDE.md) file for architecture guidelines and development instructions.

### Development Setup

1. Clone the repository:

   ```bash
   git clone https://github.com/mattermost/pluginctl.git
   cd pluginctl
   ```

2. Set up development environment (installs pinned tool versions):

   ```bash
   make dev-setup
   ```

3. Install dependencies:

   ```bash
   make deps
   ```

4. Build the project:

   ```bash
   make build
   ```

5. Test with a sample plugin:
   ```bash
   ./pluginctl info
   ```

### Development Workflow

Use these Make targets for efficient development:

```bash
# Quick development build (format, lint, build)
make dev

# Check all changes before committing (lint, security, test)
make check-changes

# Full verification (clean, lint, test, build)
make verify

# Run tests with coverage
make test-coverage

# Build for all platforms
make build-all
```

### Available Make Targets

**Development:**

- `make dev` - Quick development build
- `make check-changes` - Validate changes (lint, security, test)
- `make verify` - Full build verification
- `make fmt` - Format code
- `make clean` - Clean build artifacts

**Testing:**

- `make test` - Run tests
- `make test-coverage` - Run tests with coverage report
- `make lint` - Run linter
- `make lint-fix` - Fix linting issues automatically
- `make security` - Run security scan

**Building:**

- `make build` - Build for current platform
- `make build-all` - Build for all platforms
- `make install` - Install to GOPATH/bin

**Release:**

- `make release` - Create production release
- `make snapshot` - Create snapshot release

**Utilities:**

- `make help` - Show all available targets
- `make version` - Show version information
- `make dev-setup` - Install development tools

### Adding New Commands

1. Create a new command file (e.g., `build.go`)
2. Implement the command following the patterns in `info.go`
3. Register the command in `cmd/pluginctl/main.go`
4. Update the help text and documentation

See [CLAUDE.md](CLAUDE.md) for detailed architecture guidelines.

### Code Style

- Follow Go best practices and conventions
- Use the official Mattermost model types from `github.com/mattermost/mattermost/server/public/model`
- Maintain separation between CLI framework and command implementation
- Include comprehensive error handling with descriptive messages

### Testing

Test your changes with various plugin configurations:

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Test CLI functionality
./pluginctl info

# Test with command-line flag
./pluginctl --plugin-path /path/to/plugin info

# Test with environment variable
export PLUGINCTL_PLUGIN_PATH=/path/to/plugin
./pluginctl info

# Validate all changes before committing
make check-changes
```

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.

## Related Projects

- [Mattermost Plugin Developer Documentation](https://developers.mattermost.com/integrate/plugins/)
- [Mattermost Plugin Starter Template](https://github.com/mattermost/mattermost-plugin-starter-template)
- [Mattermost Server](https://github.com/mattermost/mattermost-server)

## Support

For questions, issues, or feature requests, please:

1. Check the [issues](https://github.com/mattermost/pluginctl/issues) page
2. Create a new issue if your problem isn't already reported
3. Join the [Mattermost Community](https://community.mattermost.com/) for general discussion
