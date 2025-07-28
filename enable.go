package pluginctl

import (
	"context"
	"fmt"

	"github.com/mattermost/mattermost/server/public/model"
)

func RunEnableCommand(args []string, pluginPath string) error {
	helpText := `Enable plugin in Mattermost server

Usage:
  pluginctl enable [options]

Options:
  --help, -h    Show this help message

Description:
  Enables the plugin in the connected Mattermost server. The plugin must already
  be uploaded to the server for this command to work.

Examples:
  pluginctl enable                            # Enable plugin from current directory
  pluginctl --plugin-path /path/to/plugin enable # Enable plugin at specific path`

	// Check for help flag
	if CheckForHelpFlag(args, helpText) {
		return nil
	}

	return runPluginCommand(args, pluginPath, enablePlugin)
}

func enablePlugin(ctx context.Context, client *model.Client4, pluginID string) error {
	Logger.Info("Enabling plugin")
	_, err := client.EnablePlugin(ctx, pluginID)
	if err != nil {
		return fmt.Errorf("failed to enable plugin: %w", err)
	}

	return nil
}
