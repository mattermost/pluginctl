package pluginctl

import (
	"context"
	"fmt"

	"github.com/mattermost/mattermost/server/public/model"
)

func RunDisableCommand(args []string, pluginPath string) error {
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
