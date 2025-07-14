package pluginctl

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mattermost/mattermost/server/public/model"
)

const PluginManifestName = "plugin.json"

// PluginCtlConfig represents the configuration for pluginctl stored in the manifest props.
type PluginCtlConfig struct {
	IgnoreAssets []string `json:"ignore_assets,omitempty"`
}

// LoadPluginManifest loads and parses the plugin.json file from the current directory.
func LoadPluginManifest() (*model.Manifest, error) {
	return LoadPluginManifestFromPath(".")
}

// LoadPluginManifestFromPath loads and parses the plugin.json file from the specified directory.
func LoadPluginManifestFromPath(dir string) (*model.Manifest, error) {
	manifestPath := filepath.Join(dir, PluginManifestName)

	// Check if plugin.json exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin.json not found in directory %s", dir)
	}

	// Read the file
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin.json: %w", err)
	}

	// Parse JSON into Manifest struct
	var manifest model.Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse plugin.json: %w", err)
	}

	return &manifest, nil
}

// HasServerCode checks if the plugin contains server-side code.
func HasServerCode(manifest *model.Manifest) bool {
	return manifest.Server != nil && len(manifest.Server.Executables) > 0
}

// HasWebappCode checks if the plugin contains webapp code.
func HasWebappCode(manifest *model.Manifest) bool {
	return manifest.Webapp != nil && manifest.Webapp.BundlePath != ""
}

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
