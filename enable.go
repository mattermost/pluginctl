package pluginctl

import (
	"context"
	"fmt"
	"log"

	"github.com/mattermost/mattermost/server/public/model"
)

func RunEnableCommand(args []string, pluginPath string) error {
	return runPluginCommand(args, pluginPath, enablePlugin)
}

func enablePlugin(ctx context.Context, client *model.Client4, pluginID string) error {
	log.Print("Enabling plugin.")
	_, err := client.EnablePlugin(ctx, pluginID)
	if err != nil {
		return fmt.Errorf("failed to enable plugin: %w", err)
	}

	return nil
}
