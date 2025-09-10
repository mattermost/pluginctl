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
```

## Usage

### Basic Commands

```bash
# Display plugin information
pluginctl info

# Show help and available commands
pluginctl --help

# Show version
pluginctl version
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

### Available Commands

Run `pluginctl --help` to see all available commands and options.

## Development

### Quick Start

```bash
# Set up development environment
make dev-setup

# Install dependencies
make deps

# Quick development build (format, lint, build)
make dev

# Run tests and checks before committing
make check-changes
```

See `make help` for the complete list of available targets.

### Adding New Commands

1. Create a new command file (e.g., `build.go`) in the root directory
2. Implement the command following existing patterns
3. Register the command in `cmd/pluginctl/main.go`
4. Update the help text

See [CLAUDE.md](CLAUDE.md) for detailed architecture guidelines.

## Contributing

We welcome contributions to `pluginctl`! Please see the [CLAUDE.md](CLAUDE.md) file for architecture guidelines and development patterns.

### Code Style

- Follow Go best practices and conventions
- Use the official Mattermost model types from `github.com/mattermost/mattermost/server/public/model`
- Maintain separation between CLI framework and command implementation
- Include comprehensive error handling with descriptive messages

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.

## Related Projects

- [Mattermost Plugin Developer Documentation](https://developers.mattermost.com/integrate/plugins/)
- [Mattermost Plugin Starter Template](https://github.com/mattermost/mattermost-plugin-starter-template)
- [Mattermost Server](https://github.com/mattermost/mattermost-server)

## Support

For questions, issues, or feature requests, please:

- Check the [issues](https://github.com/mattermost/pluginctl/issues) or create a new issue if your problem isn't already reported
- Join the [Mattermost Community](https://community.mattermost.com/) for general discussion
