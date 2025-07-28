package pluginctl

import (
	"context"
	"fmt"

	"github.com/mattermost/mattermost/server/public/model"
)

func RunDisableCommand(args []string, pluginPath string) error {
	helpText := `Disable plugin in Mattermost server

Usage:
  pluginctl disable [options]

Options:
  --help, -h    Show this help message

Description:
  Disables the plugin in the connected Mattermost server. The plugin will
  remain uploaded but will be inactive.

Examples:
  pluginctl disable                           # Disable plugin from current directory
  pluginctl --plugin-path /path/to/plugin disable # Disable plugin at specific path`

	// Check for help flag
	if CheckForHelpFlag(args, helpText) {
		return nil
	}

	return runPluginCommand(args, pluginPath, disablePlugin)
}

func disablePlugin(ctx context.Context, client *model.Client4, pluginID string) error {
	Logger.Info("Disabling plugin")
	_, err := client.DisablePlugin(ctx, pluginID)
	if err != nil {
		return fmt.Errorf("failed to disable plugin: %w", err)
	}

	return nil
}
