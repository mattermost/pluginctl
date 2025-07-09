package pluginctl

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed assets/**/*
var assetsFS embed.FS

func RunUpdateAssetsCommand(args []string, pluginPath string) error {
	if len(args) > 0 {
		return fmt.Errorf("updateassets command does not accept arguments")
	}

	fmt.Printf("Updating assets in plugin directory: %s\n", pluginPath)

	// Load plugin manifest to check for webapp code
	manifest, err := LoadPluginManifestFromPath(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to load plugin manifest: %w", err)
	}

	// Check if the plugin has webapp code according to manifest
	hasWebapp := HasWebappCode(manifest)

	// Counter for updated files
	var updatedCount int

	// Walk through the embedded assets
	err = fs.WalkDir(assetsFS, "assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root assets directory
		if path == "assets" {
			return nil
		}

		// Remove the "assets/" prefix to get the relative path
		relativePath := path[7:] // len("assets/") = 7

		// Skip webapp assets if plugin doesn't have webapp code
		if !hasWebapp && strings.HasPrefix(relativePath, "webapp") {
			return nil
		}

		targetPath := filepath.Join(pluginPath, relativePath)

		if d.IsDir() {
			// Create directory if it doesn't exist
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}
		} else {
			// Read file content from embedded FS
			content, err := assetsFS.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read embedded file %s: %w", path, err)
			}

			// Check if target file exists and compare content
			existingContent, err := os.ReadFile(targetPath)
			if err == nil && bytes.Equal(existingContent, content) {
				// File exists and content is identical, skip update
				return nil
			}

			// Create parent directory if it doesn't exist
			parentDir := filepath.Dir(targetPath)
			if err := os.MkdirAll(parentDir, 0755); err != nil {
				return fmt.Errorf("failed to create parent directory %s: %w", parentDir, err)
			}

			// Write file to target location
			if err := os.WriteFile(targetPath, content, 0644); err != nil {
				return fmt.Errorf("failed to write file %s: %w", targetPath, err)
			}
			fmt.Printf("Updated file: %s\n", relativePath)
			updatedCount++
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to update assets: %w", err)
	}

	fmt.Printf("Assets updated successfully! (%d files updated)\n", updatedCount)
	return nil
}
