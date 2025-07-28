package pluginctl

import (
	"context"

	"github.com/mattermost/mattermost/server/public/model"
)

func RunResetCommand(args []string, pluginPath string) error {
	helpText := `Reset plugin (disable then enable)

Usage:
  pluginctl reset [options]

Options:
  --help, -h    Show this help message

Description:
  Resets the plugin by first disabling it and then enabling it. This is useful
  for restarting a plugin without having to redeploy it.

Examples:
  pluginctl reset                             # Reset plugin from current directory
  pluginctl --plugin-path /path/to/plugin reset # Reset plugin at specific path`

	// Check for help flag
	if CheckForHelpFlag(args, helpText) {
		return nil
	}

	return runPluginCommand(args, pluginPath, resetPlugin)
}

func resetPlugin(ctx context.Context, client *model.Client4, pluginID string) error {
	err := disablePlugin(ctx, client, pluginID)
	if err != nil {
		return err
	}

	err = enablePlugin(ctx, client, pluginID)
	if err != nil {
		return err
	}

	return nil
}
