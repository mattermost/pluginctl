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

//go:embed assets/*
var assetsFS embed.FS

const (
	assetsPrefix         = "assets/"
	assetsPrefixLen      = 7
	directoryPermissions = 0o750
	filePermissions      = 0o600
)

func RunUpdateAssetsCommand(args []string, pluginPath string) error {
	if len(args) > 0 {
		return fmt.Errorf("updateassets command does not accept arguments")
	}

	Logger.Info("Updating assets in plugin directory", "path", pluginPath)

	manifest, err := LoadPluginManifestFromPath(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to load plugin manifest: %w", err)
	}

	hasWebapp := HasWebappCode(manifest)
	updatedCount := 0

	config := AssetProcessorConfig{
		pluginPath:   pluginPath,
		hasWebapp:    hasWebapp,
		updatedCount: &updatedCount,
	}

	err = fs.WalkDir(assetsFS, "assets", func(path string, d fs.DirEntry, err error) error {
		return processAssetEntry(path, d, err, config)
	})

	if err != nil {
		return fmt.Errorf("failed to update assets: %w", err)
	}

	Logger.Info("Assets updated successfully!", "files_updated", updatedCount)

	return nil
}

type AssetProcessorConfig struct {
	pluginPath   string
	hasWebapp    bool
	updatedCount *int
}

func processAssetEntry(path string, d fs.DirEntry, err error, config AssetProcessorConfig) error {
	if err != nil {
		return err
	}

	if path == "assets" {
		return nil
	}

	relativePath := path[assetsPrefixLen:]

	if !config.hasWebapp && strings.HasPrefix(relativePath, "webapp") {
		return nil
	}

	targetPath := filepath.Join(config.pluginPath, relativePath)

	if d.IsDir() {
		return createDirectory(targetPath)
	}

	return processAssetFile(path, targetPath, relativePath, config.updatedCount)
}

func processAssetFile(embeddedPath, targetPath, relativePath string, updatedCount *int) error {
	shouldUpdate, err := shouldUpdateFile(embeddedPath, targetPath)
	if err != nil {
		return err
	}

	if shouldUpdate {
		err = updateFile(embeddedPath, targetPath, relativePath)
		if err != nil {
			return err
		}
		(*updatedCount)++
	}

	return nil
}

func createDirectory(targetPath string) error {
	if err := os.MkdirAll(targetPath, directoryPermissions); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
	}

	return nil
}

func shouldUpdateFile(embeddedPath, targetPath string) (bool, error) {
	content, err := assetsFS.ReadFile(embeddedPath)
	if err != nil {
		return false, fmt.Errorf("failed to read embedded file %s: %w", embeddedPath, err)
	}

	existingContent, err := os.ReadFile(targetPath)
	if err != nil {
		// File doesn't exist or other error, should update
		return true, nil //nolint:nilerr
	}

	return !bytes.Equal(existingContent, content), nil
}

func updateFile(embeddedPath, targetPath, relativePath string) error {
	content, err := assetsFS.ReadFile(embeddedPath)
	if err != nil {
		return fmt.Errorf("failed to read embedded file %s: %w", embeddedPath, err)
	}

	parentDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(parentDir, directoryPermissions); err != nil {
		return fmt.Errorf("failed to create parent directory %s: %w", parentDir, err)
	}

	if err := os.WriteFile(targetPath, content, filePermissions); err != nil {
		return fmt.Errorf("failed to write file %s: %w", targetPath, err)
	}

	Logger.Info("Updated file", "path", relativePath)

	return nil
}
