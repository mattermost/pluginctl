package pluginctl

import (
	"fmt"
	"runtime/debug"
)

// RunVersionCommand implements the 'version' command functionality.
func RunVersionCommand(args []string) error {
	version := GetVersion()
	fmt.Printf("pluginctl version %s\n", version)

	return nil
}

// GetVersion returns the version information from build info.
func GetVersion() string {
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
			const shortHashLength = 7
			if len(setting.Value) >= shortHashLength {
				return setting.Value[:shortHashLength]
			}

			return setting.Value
		}
	}

	return "dev"
}
