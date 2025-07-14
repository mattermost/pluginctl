package pluginctl

import (
	"fmt"
	"path/filepath"
)

// RunManifestCommand implements the 'manifest' command functionality with subcommands.
func RunManifestCommand(args []string, pluginPath string) error {
	if len(args) == 0 {
		return fmt.Errorf("manifest command requires a subcommand: id, version, has_server, has_webapp, check")
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Load plugin manifest
	manifest, err := LoadPluginManifestFromPath(absPath)
	if err != nil {
		return fmt.Errorf("failed to load plugin manifest: %w", err)
	}

	subcommand := args[0]
	switch subcommand {
	case "id":
		fmt.Println(manifest.Id)
	case "version":
		fmt.Println(manifest.Version)
	case "has_server":
		if HasServerCode(manifest) {
			fmt.Println("true")
		} else {
			fmt.Println("false")
		}
	case "has_webapp":
		if HasWebappCode(manifest) {
			fmt.Println("true")
		} else {
			fmt.Println("false")
		}
	case "check":
		if err := manifest.IsValid(); err != nil {
			Logger.Error("Plugin manifest validation failed", "error", err)

			return err
		}
		Logger.Info("Plugin manifest is valid")
	default:
		return fmt.Errorf("unknown subcommand: %s. Available subcommands: id, version, has_server, has_webapp, check",
			subcommand)
	}

	return nil
}
