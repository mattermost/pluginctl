package pluginctl

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
)

const commandTimeout = 120 * time.Second

func getClient(ctx context.Context) (*model.Client4, error) {
	socketPath := os.Getenv("MM_LOCALSOCKETPATH")
	if socketPath == "" {
		socketPath = model.LocalModeSocketPath
	}

	client, connected := getUnixClient(socketPath)
	if connected {
		Logger.Info("Connecting using local mode", "socket_path", socketPath)

		return client, nil
	}

	if os.Getenv("MM_LOCALSOCKETPATH") != "" {
		Logger.Info("No socket found for local mode deployment. Attempting to authenticate with credentials.", "socket_path", socketPath)
	}

	siteURL := os.Getenv("MM_SERVICESETTINGS_SITEURL")
	adminToken := os.Getenv("MM_ADMIN_TOKEN")
	adminUsername := os.Getenv("MM_ADMIN_USERNAME")
	adminPassword := os.Getenv("MM_ADMIN_PASSWORD")

	if siteURL == "" {
		return nil, errors.New("MM_SERVICESETTINGS_SITEURL is not set")
	}

	client = model.NewAPIv4Client(siteURL)

	if adminToken != "" {
		Logger.Info("Authenticating using token", "site_url", siteURL)
		client.SetToken(adminToken)

		return client, nil
	}

	if adminUsername != "" && adminPassword != "" {
		client := model.NewAPIv4Client(siteURL)
		Logger.Info("Authenticating with credentials", "username", adminUsername, "site_url", siteURL)
		_, _, err := client.Login(ctx, adminUsername, adminPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to login as %s: %w", adminUsername, err)
		}

		return client, nil
	}

	return nil, errors.New("one of MM_ADMIN_TOKEN or MM_ADMIN_USERNAME/MM_ADMIN_PASSWORD must be defined")
}

func getUnixClient(socketPath string) (*model.Client4, bool) {
	_, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, false
	}

	return model.NewAPIv4SocketClient(socketPath), true
}

// runPluginCommand executes a plugin command using the plugin from the current folder.
func runPluginCommand(
	_ []string,
	pluginPath string,
	action func(context.Context, *model.Client4, string) error,
) error {
	// Load plugin manifest to get the plugin ID
	manifest, err := LoadPluginManifestFromPath(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to load plugin manifest: %w", err)
	}

	pluginID := manifest.Id
	if pluginID == "" {
		return errors.New("plugin ID not found in manifest")
	}

	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	return action(ctx, client, pluginID)
}
