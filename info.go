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
	fmt.Printf("Plugin Information:\n")
	fmt.Printf("==================\n\n")

	// Basic plugin info
	fmt.Printf("ID:              %s\n", manifest.Id)
	fmt.Printf("Name:            %s\n", manifest.Name)
	fmt.Printf("Version:         %s\n", manifest.Version)

	// Minimum Mattermost version
	if manifest.MinServerVersion != "" {
		fmt.Printf("Min MM Version:  %s\n", manifest.MinServerVersion)
	} else {
		fmt.Printf("Min MM Version:  Not specified\n")
	}

	// Description if available
	if manifest.Description != "" {
		fmt.Printf("Description:     %s\n", manifest.Description)
	}

	fmt.Printf("\nCode Components:\n")
	fmt.Printf("================\n")

	// Server code presence
	if HasServerCode(manifest) {
		fmt.Printf("Server Code:     Yes\n")
		if manifest.Server != nil && len(manifest.Server.Executables) > 0 {
			fmt.Printf("  Executables:   ")
			first := true
			for platform := range manifest.Server.Executables {
				if !first {
					fmt.Printf(", ")
				}
				fmt.Printf("%s", platform)
				first = false
			}
			fmt.Printf("\n")
		}
	} else {
		fmt.Printf("Server Code:     No\n")
	}

	// Webapp code presence
	if HasWebappCode(manifest) {
		fmt.Printf("Webapp Code:     Yes\n")
		if manifest.Webapp != nil && manifest.Webapp.BundlePath != "" {
			fmt.Printf("  Bundle Path:   %s\n", manifest.Webapp.BundlePath)
		}
	} else {
		fmt.Printf("Webapp Code:     No\n")
	}

	// Settings schema
	if manifest.SettingsSchema != nil {
		fmt.Printf("Settings Schema: Yes\n")
	} else {
		fmt.Printf("Settings Schema: No\n")
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
