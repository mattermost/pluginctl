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
	fmt.Println("Plugin Information:")
	fmt.Println("==================")

	fmt.Printf("ID:              %s\n", manifest.Id)
	fmt.Printf("Name:            %s\n", manifest.Name)
	fmt.Printf("Version:         %s\n", manifest.Version)

	minVersion := manifest.MinServerVersion
	if minVersion == "" {
		minVersion = "Not specified"
	}
	fmt.Printf("Min MM Version:  %s\n", minVersion)

	if manifest.Description != "" {
		fmt.Printf("Description:     %s\n", manifest.Description)
	}
}

// printCodeComponents prints information about server and webapp code.
func printCodeComponents(manifest *model.Manifest) {
	fmt.Println("Code Components:")
	fmt.Println("================")

	printServerCodeInfo(manifest)
	printWebappCodeInfo(manifest)
}

// printServerCodeInfo prints server code information.
func printServerCodeInfo(manifest *model.Manifest) {
	if HasServerCode(manifest) {
		fmt.Println("Server Code:     Yes")
		if manifest.Server != nil && len(manifest.Server.Executables) > 0 {
			fmt.Print("Executables:     ")
			first := true
			for platform := range manifest.Server.Executables {
				if !first {
					fmt.Print(", ")
				}
				fmt.Print(platform)
				first = false
			}
			fmt.Println()
		}
	} else {
		fmt.Println("Server Code:     No")
	}
}

// printWebappCodeInfo prints webapp code information.
func printWebappCodeInfo(manifest *model.Manifest) {
	if HasWebappCode(manifest) {
		fmt.Println("Webapp Code:     Yes")
		if manifest.Webapp != nil && manifest.Webapp.BundlePath != "" {
			fmt.Printf("Bundle Path:   %s\n", manifest.Webapp.BundlePath)
		}
	} else {
		fmt.Println("Webapp Code:     No")
	}
}

// printSettingsSchema prints settings schema information.
func printSettingsSchema(manifest *model.Manifest) {
	value := "No"
	if manifest.SettingsSchema != nil {
		value = "Yes"
	}
	fmt.Printf("Settings Schema: %s\n", value)
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
