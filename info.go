package pluginctl

import (
	"fmt"
	"path/filepath"

	"github.com/mattermost/mattermost/server/public/model"
)

// InfoCommand implements the 'info' command functionality.
func InfoCommand() error {
	manifest, err := LoadPluginManifest()
	if err != nil {
		return fmt.Errorf("failed to load plugin manifest: %w", err)
	}

	return PrintPluginInfo(manifest)
}

// PrintPluginInfo displays formatted plugin information.
func PrintPluginInfo(manifest *model.Manifest) error {
	Logger.Info("Plugin Information:")
	Logger.Info("==================")

	// Basic plugin info
	Logger.Info("ID:", "value", manifest.Id)
	Logger.Info("Name:", "value", manifest.Name)
	Logger.Info("Version:", "value", manifest.Version)

	// Minimum Mattermost version
	if manifest.MinServerVersion != "" {
		Logger.Info("Min MM Version:", "value", manifest.MinServerVersion)
	} else {
		Logger.Info("Min MM Version:", "value", "Not specified")
	}

	// Description if available
	if manifest.Description != "" {
		Logger.Info("Description:", "value", manifest.Description)
	}

	Logger.Info("Code Components:")
	Logger.Info("================")

	// Server code presence
	if HasServerCode(manifest) {
		Logger.Info("Server Code:", "value", "Yes")
		if manifest.Server != nil && len(manifest.Server.Executables) > 0 {
			var executables []string
			for platform := range manifest.Server.Executables {
				executables = append(executables, platform)
			}
			Logger.Info("Executables:", "platforms", executables)
		}
	} else {
		Logger.Info("Server Code:", "value", "No")
	}

	// Webapp code presence
	if HasWebappCode(manifest) {
		Logger.Info("Webapp Code:", "value", "Yes")
		if manifest.Webapp != nil && manifest.Webapp.BundlePath != "" {
			Logger.Info("Bundle Path:", "value", manifest.Webapp.BundlePath)
		}
	} else {
		Logger.Info("Webapp Code:", "value", "No")
	}

	// Settings schema
	if manifest.SettingsSchema != nil {
		Logger.Info("Settings Schema:", "value", "Yes")
	} else {
		Logger.Info("Settings Schema:", "value", "No")
	}

	return nil
}

// InfoCommandWithPath implements the 'info' command with a custom path.
func InfoCommandWithPath(path string) error {
	manifest, err := LoadPluginManifestFromPath(path)
	if err != nil {
		return fmt.Errorf("failed to load plugin manifest from %s: %w", path, err)
	}

	return PrintPluginInfo(manifest)
}

// RunInfoCommand implements the 'info' command functionality with plugin path.
func RunInfoCommand(args []string, pluginPath string) error {
	// Convert to absolute path
	absPath, err := filepath.Abs(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	return InfoCommandWithPath(absPath)
}
