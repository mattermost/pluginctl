package pluginctl

import (
	"context"
	"fmt"

	"github.com/mattermost/mattermost/server/public/model"
)

func RunEnableCommand(args []string, pluginPath string) error {
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
