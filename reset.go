package pluginctl

import (
	"context"

	"github.com/mattermost/mattermost/server/public/model"
)

func RunResetCommand(args []string, pluginPath string) error {
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
