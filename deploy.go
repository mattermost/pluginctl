package pluginctl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mattermost/mattermost/server/public/model"
)

func RunDeployCommand(args []string, pluginPath string) error {
	var bundlePath string

	// Parse flags
	i := 0
	for i < len(args) {
		switch args[i] {
		case "--bundle-path":
			if i+1 >= len(args) {
				return fmt.Errorf("--bundle-path flag requires a value")
			}
			bundlePath = args[i+1]
			i += 2
		default:
			i++
		}
	}

	// If no bundle path provided, auto-discover from dist folder
	if bundlePath == "" {
		manifest, err := LoadPluginManifestFromPath(pluginPath)
		if err != nil {
			return fmt.Errorf("failed to load plugin manifest: %w", err)
		}

		expectedBundleName := fmt.Sprintf("%s-%s.tar.gz", manifest.Id, manifest.Version)
		bundlePath = filepath.Join(pluginPath, "dist", expectedBundleName)

		if _, err := os.Stat(bundlePath); os.IsNotExist(err) {
			return fmt.Errorf("bundle not found at %s - run 'make bundle' to build the plugin first", bundlePath)
		}
	}

	// Validate bundle file exists
	if _, err := os.Stat(bundlePath); os.IsNotExist(err) {
		return fmt.Errorf("bundle file not found: %s", bundlePath)
	}

	// Load manifest to get plugin ID
	manifest, err := LoadPluginManifestFromPath(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to load plugin manifest: %w", err)
	}

	pluginID := manifest.Id
	if pluginID == "" {
		return fmt.Errorf("plugin ID not found in manifest")
	}

	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	return deployPlugin(ctx, client, pluginID, bundlePath)
}

func deployPlugin(ctx context.Context, client *model.Client4, pluginID, bundlePath string) error {
	pluginBundle, err := os.Open(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to open bundle file %s: %w", bundlePath, err)
	}
	defer func() {
		if closeErr := pluginBundle.Close(); closeErr != nil {
			Logger.Error("Failed to close plugin bundle", "error", closeErr)
		}
	}()

	Logger.Info("Uploading plugin bundle", "bundle_path", bundlePath)
	_, _, err = client.UploadPluginForced(ctx, pluginBundle)
	if err != nil {
		return fmt.Errorf("failed to upload plugin bundle: %w", err)
	}

	Logger.Info("Enabling plugin", "plugin_id", pluginID)
	_, err = client.EnablePlugin(ctx, pluginID)
	if err != nil {
		return fmt.Errorf("failed to enable plugin: %w", err)
	}

	Logger.Info("Plugin deployed successfully", "plugin_id", pluginID)

	return nil
}
