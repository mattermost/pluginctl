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
	printBasicInfo(manifest)
	printCodeComponents(manifest)
	printSettingsSchema(manifest)

	return nil
}

// printBasicInfo prints basic plugin information.
func printBasicInfo(manifest *model.Manifest) {
	Logger.Info("Plugin Information:")
	Logger.Info("==================")

	Logger.Info("ID:", "value", manifest.Id)
	Logger.Info("Name:", "value", manifest.Name)
	Logger.Info("Version:", "value", manifest.Version)

	minVersion := manifest.MinServerVersion
	if minVersion == "" {
		minVersion = "Not specified"
	}
	Logger.Info("Min MM Version:", "value", minVersion)

	if manifest.Description != "" {
		Logger.Info("Description:", "value", manifest.Description)
	}
}

// printCodeComponents prints information about server and webapp code.
func printCodeComponents(manifest *model.Manifest) {
	Logger.Info("Code Components:")
	Logger.Info("================")

	printServerCodeInfo(manifest)
	printWebappCodeInfo(manifest)
}

// printServerCodeInfo prints server code information.
func printServerCodeInfo(manifest *model.Manifest) {
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
}

// printWebappCodeInfo prints webapp code information.
func printWebappCodeInfo(manifest *model.Manifest) {
	if HasWebappCode(manifest) {
		Logger.Info("Webapp Code:", "value", "Yes")
		if manifest.Webapp != nil && manifest.Webapp.BundlePath != "" {
			Logger.Info("Bundle Path:", "value", manifest.Webapp.BundlePath)
		}
	} else {
		Logger.Info("Webapp Code:", "value", "No")
	}
}

// printSettingsSchema prints settings schema information.
func printSettingsSchema(manifest *model.Manifest) {
	value := "No"
	if manifest.SettingsSchema != nil {
		value = "Yes"
	}
	Logger.Info("Settings Schema:", "value", value)
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
