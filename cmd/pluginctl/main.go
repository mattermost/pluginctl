package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/mattermost/mattermost/server/public/model"
)

const (
	ExitSuccess   = 0
	ExitError     = 1
	EnvPluginPath = "PLUGINCTL_PLUGIN_PATH"
)

func main() {
	var pluginPath string

	flag.StringVar(&pluginPath, "plugin-path", "", "Path to plugin directory (overrides PLUGINCTL_PLUGIN_PATH)")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No command specified\n\n")
		showUsage()
		os.Exit(ExitError)
	}

	command := args[0]
	commandArgs := args[1:]

	// Determine plugin path from flag, environment variable, or current directory
	effectivePluginPath := getEffectivePluginPath(pluginPath)

	if err := runCommand(command, commandArgs, effectivePluginPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(ExitError)
	}
}

func runCommand(command string, args []string, pluginPath string) error {
	switch command {
	case "info":
		return runInfoCommand(args, pluginPath)
	case "help":
		showUsage()

		return nil
	case "version":
		return runVersionCommand(args)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func runInfoCommand(args []string, pluginPath string) error {
	// Convert to absolute path
	absPath, err := filepath.Abs(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	return infoCommandWithPath(absPath)
}

func runVersionCommand(args []string) error {
	version := getVersion()
	fmt.Printf("pluginctl version %s\n", version)

	return nil
}

// getVersion returns the version information from build info.
func getVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	// First try to get version from main module
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}

	// Look for version in build settings (set by goreleaser)
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			// Return short commit hash if no version tag
			if len(setting.Value) >= 7 {
				return setting.Value[:7]
			}
			return setting.Value
		}
	}

	return "dev"
}

// getEffectivePluginPath determines the plugin path from flag, environment variable, or current directory.
func getEffectivePluginPath(flagPath string) string {
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

func infoCommandWithPath(path string) error {
	manifest, err := loadPluginManifestFromPath(path)
	if err != nil {
		return fmt.Errorf("failed to load plugin manifest from %s: %w", path, err)
	}

	return printPluginInfo(manifest)
}

func loadPluginManifestFromPath(dir string) (*model.Manifest, error) {
	manifestPath := filepath.Join(dir, "plugin.json")

	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin.json not found in directory %s", dir)
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin.json: %w", err)
	}

	var manifest model.Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse plugin.json: %w", err)
	}

	return &manifest, nil
}

func printPluginInfo(manifest *model.Manifest) error {
	fmt.Printf("Plugin Information:\n")
	fmt.Printf("==================\n\n")

	fmt.Printf("ID:              %s\n", manifest.Id)
	fmt.Printf("Name:            %s\n", manifest.Name)
	fmt.Printf("Version:         %s\n", manifest.Version)

	if manifest.MinServerVersion != "" {
		fmt.Printf("Min MM Version:  %s\n", manifest.MinServerVersion)
	} else {
		fmt.Printf("Min MM Version:  Not specified\n")
	}

	if manifest.Description != "" {
		fmt.Printf("Description:     %s\n", manifest.Description)
	}

	fmt.Printf("\nCode Components:\n")
	fmt.Printf("================\n")

	if hasServerCode(manifest) {
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

	if hasWebappCode(manifest) {
		fmt.Printf("Webapp Code:     Yes\n")
		if manifest.Webapp != nil && manifest.Webapp.BundlePath != "" {
			fmt.Printf("  Bundle Path:   %s\n", manifest.Webapp.BundlePath)
		}
	} else {
		fmt.Printf("Webapp Code:     No\n")
	}

	if manifest.SettingsSchema != nil {
		fmt.Printf("Settings Schema: Yes\n")
	} else {
		fmt.Printf("Settings Schema: No\n")
	}

	return nil
}

func hasServerCode(manifest *model.Manifest) bool {
	return manifest.Server != nil && len(manifest.Server.Executables) > 0
}

func hasWebappCode(manifest *model.Manifest) bool {
	return manifest.Webapp != nil && manifest.Webapp.BundlePath != ""
}

func showUsage() {
	fmt.Printf(`pluginctl - Mattermost Plugin Development CLI

Usage:
  pluginctl [global options] <command> [command options] [arguments...]

Global Options:
  --plugin-path PATH   Path to plugin directory (overrides PLUGINCTL_PLUGIN_PATH)

Commands:
  info           Display plugin information
  help           Show this help message
  version        Show version information

Examples:
  pluginctl info                              # Show info for plugin in current directory
  pluginctl --plugin-path /path/to/plugin info # Show info for plugin at specific path
  export PLUGINCTL_PLUGIN_PATH=/path/to/plugin
  pluginctl info                              # Show info using environment variable
  pluginctl version                           # Show version information

Environment Variables:
  PLUGINCTL_PLUGIN_PATH   Default plugin directory path

For more information about Mattermost plugin development, visit:
https://developers.mattermost.com/integrate/plugins/
`)
}
