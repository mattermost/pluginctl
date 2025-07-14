package pluginctl

import (
	"fmt"
	"path/filepath"
)

// RunManifestCommand implements the 'manifest' command functionality with subcommands.
func RunManifestCommand(args []string, pluginPath string) error {
	if len(args) == 0 {
		return fmt.Errorf("manifest command requires a subcommand: id, version, has_server, has_webapp")
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
	default:
		return fmt.Errorf("unknown subcommand: %s. Available subcommands: id, version, has_server, has_webapp",
			subcommand)
	}

	return nil
}
