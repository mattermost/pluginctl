package pluginctl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mattermost/mattermost/server/public/model"
)

func RunDeployCommand(args []string, pluginPath string) error {
	helpText := getDeployHelpText()

	// Check for help flag
	if CheckForHelpFlag(args, helpText) {
		return nil
	}

	bundlePath, err := parseDeployFlags(args, helpText)
	if err != nil {
		return err
	}

	bundlePath, err = resolveBundlePath(bundlePath, pluginPath)
	if err != nil {
		return err
	}

	pluginID, err := getPluginIDFromManifest(pluginPath)
	if err != nil {
		return err
	}

	return deployPluginBundle(pluginID, bundlePath)
}

func getDeployHelpText() string {
	return `Upload and enable plugin bundle

Usage:
  pluginctl deploy [options]

Options:
  --bundle-path PATH    Path to plugin bundle file (.tar.gz)
  --help, -h           Show this help message

Description:
  Uploads a plugin bundle to the Mattermost server and enables it. If no
  bundle path is specified, it will auto-discover the bundle from the dist/
  directory based on the plugin manifest.

Examples:
  pluginctl deploy                            # Deploy bundle from ./dist/
  pluginctl deploy --bundle-path ./bundle.tar.gz # Deploy specific bundle file
  pluginctl --plugin-path /path/to/plugin deploy # Deploy plugin at specific path`
}

func parseDeployFlags(args []string, helpText string) (string, error) {
	var bundlePath string

	i := 0
	for i < len(args) {
		switch args[i] {
		case "--bundle-path":
			if i+1 >= len(args) {
				return "", ShowErrorWithHelp("--bundle-path flag requires a value", helpText)
			}
			bundlePath = args[i+1]
			i += 2
		default:
			i++
		}
	}

	return bundlePath, nil
}

func resolveBundlePath(bundlePath, pluginPath string) (string, error) {
	// If bundle path provided, validate it exists
	if bundlePath != "" {
		if _, err := os.Stat(bundlePath); os.IsNotExist(err) {
			return "", fmt.Errorf("bundle file not found: %s", bundlePath)
		}

		return bundlePath, nil
	}

	// Auto-discover from dist folder
	manifest, err := LoadPluginManifestFromPath(pluginPath)
	if err != nil {
		return "", fmt.Errorf("failed to load plugin manifest: %w", err)
	}

	expectedBundleName := fmt.Sprintf("%s-%s.tar.gz", manifest.Id, manifest.Version)
	bundlePath = filepath.Join(pluginPath, "dist", expectedBundleName)

	if _, err := os.Stat(bundlePath); os.IsNotExist(err) {
		return "", fmt.Errorf("bundle not found at %s - run 'make bundle' to build the plugin first", bundlePath)
	}

	return bundlePath, nil
}

func getPluginIDFromManifest(pluginPath string) (string, error) {
	manifest, err := LoadPluginManifestFromPath(pluginPath)
	if err != nil {
		return "", fmt.Errorf("failed to load plugin manifest: %w", err)
	}

	pluginID := manifest.Id
	if pluginID == "" {
		return "", fmt.Errorf("plugin ID not found in manifest")
	}

	return pluginID, nil
}

func deployPluginBundle(pluginID, bundlePath string) error {
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
