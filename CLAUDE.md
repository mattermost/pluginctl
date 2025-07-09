# pluginctl - Claude Memory

## Critical Architecture Rules

### Command Structure

- **CRITICAL**: ALL command logic in separate files in ROOT folder (e.g., `info.go`, `enable.go`)
- **NEVER** put command implementation in `cmd/pluginctl/main.go` - only CLI framework code
- **Command pattern**: `run[Command]Command(args []string, pluginPath string) error`
- **Registration**: Add to `runCommand()` switch in main.go

### Dependencies & Types

- Use `github.com/mattermost/mattermost/server/public/model.Manifest`
- Commands in `pluginctl` package, main.go calls them

### Plugin Path Resolution

1. `--plugin-path` flag
2. `PLUGINCTL_PLUGIN_PATH` env var
3. Current directory (default)

### Logging

- **CRITICAL**: Use `pluginctl.Logger` (global slog instance)
- **Error**: `pluginctl.Logger.Error("message", "error", err)`
- **Info**: `pluginctl.Logger.Info("message")`
- **NEVER** use `fmt.Print*` or `log.*`

### Build & Development

- **CRITICAL**: Use `make dev` for testing builds, NOT `go build`
- **Before commit**: `make check-changes`
- **Dependencies**: `make deps && go mod tidy`

### Error Handling

- Commands return errors, main.go handles exit codes
- Use `fmt.Errorf("context: %w", err)` for wrapping

### Adding New Commands

1. Create `[command].go` in root with `Run[Command]Command` function
2. Add case to `runCommand()` switch in main.go
3. Update `showUsage()` in main.go

### Key Patterns

- Always validate plugin.json exists before operations
- Use structured logging with key-value pairs
- Follow existing naming conventions
- Keep command files self-contained and testable
