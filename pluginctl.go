package pluginctl

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mattermost/mattermost/server/public/model"
)

const (
	HelpFlagLong  = "--help"
	HelpFlagShort = "-h"

	VersionCommand = "version"
	DevVersion     = "dev"
	UnknownVersion = "unknown"
)

// IsValidPluginDirectory checks if the current directory contains a valid plugin.
func IsValidPluginDirectory() error {
	_, err := LoadPluginManifest()

	return err
}

// GetEffectivePluginPath determines the plugin path from flag, environment variable, or current directory.
func GetEffectivePluginPath(flagPath string) string {
	const EnvPluginPath = "PLUGINCTL_PLUGIN_PATH"

	// Priority: 1. Command line flag, 2. Environment variable, 3. Current directory
	if flagPath != "" {
		return flagPath
	}

	if envPath := os.Getenv(EnvPluginPath); envPath != "" {
		return envPath
	}

	// Default to current directory
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}

	return cwd
}

// ParsePluginCtlConfig extracts and parses the pluginctl configuration from the manifest props.
func ParsePluginCtlConfig(manifest *model.Manifest) (*PluginCtlConfig, error) {
	// Default configuration
	config := &PluginCtlConfig{
		IgnoreAssets: []string{},
	}

	// Check if props exist
	if manifest.Props == nil {
		return config, nil
	}

	// Check if pluginctl config exists in props
	pluginctlData, exists := manifest.Props["pluginctl"]
	if !exists {
		return config, nil
	}

	// Convert to JSON and parse
	jsonData, err := json.Marshal(pluginctlData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal pluginctl config: %w", err)
	}

	if err := json.Unmarshal(jsonData, config); err != nil {
		return nil, fmt.Errorf("failed to parse pluginctl config: %w", err)
	}

	return config, nil
}

// CheckForHelpFlag checks if --help is in the arguments and shows help if found.
// Returns true if help was shown, false otherwise.
func CheckForHelpFlag(args []string, helpText string) bool {
	for _, arg := range args {
		if arg == HelpFlagLong || arg == HelpFlagShort {
			Logger.Info(helpText)

			return true
		}
	}

	return false
}

// ShowErrorWithHelp displays an error message followed by command help.
func ShowErrorWithHelp(errorMsg, helpText string) error {
	Logger.Error(errorMsg)
	Logger.Info(helpText)

	return fmt.Errorf("%s", errorMsg)
}

// SavePluginCtlConfig saves the pluginctl configuration to the manifest props.
func SavePluginCtlConfig(manifest *model.Manifest, config *PluginCtlConfig) {
	if manifest.Props == nil {
		manifest.Props = make(map[string]interface{})
	}

	manifest.Props["pluginctl"] = config
}

// ValidateAndUpdateVersion checks the plugin version and updates it if necessary.
func ValidateAndUpdateVersion(pluginPath string) error {
	// Load the manifest
	manifest, err := LoadPluginManifestFromPath(pluginPath)
	if err != nil {
		// If there's no plugin.json, skip version validation
		Logger.Debug("No plugin.json found, skipping version validation", "error", err)

		return nil
	}

	// Get current pluginctl version
	currentVersion := GetVersion()

	// Parse existing pluginctl config
	config, err := ParsePluginCtlConfig(manifest)
	if err != nil {
		return fmt.Errorf("failed to parse pluginctl config: %w", err)
	}

	// Check if stored version is higher than current version
	if config.Version != "" && isVersionHigher(config.Version, currentVersion) {
		return fmt.Errorf("plugin was last modified with pluginctl version %s, "+
			"which is higher than current version %s. Please upgrade pluginctl",
			config.Version, currentVersion)
	}

	// Update version if different
	if config.Version != currentVersion {
		config.Version = currentVersion
		SavePluginCtlConfig(manifest, config)

		// Save the updated manifest
		if err := WritePluginManifest(manifest, pluginPath); err != nil {
			return fmt.Errorf("failed to update version in manifest: %w", err)
		}
	}

	return nil
}

// isVersionHigher compares two version strings and returns true if the stored version is higher
// than the current version.
func isVersionHigher(storedVersion, currentVersion string) bool {
	// Handle special cases
	if storedVersion == currentVersion {
		return false
	}
	if currentVersion == DevVersion || currentVersion == UnknownVersion {
		return false
	}
	if storedVersion == DevVersion || storedVersion == UnknownVersion {
		return false
	}

	// Simple string comparison for versions
	// This is a basic implementation - in production you might want semantic versioning
	return storedVersion > currentVersion
}
