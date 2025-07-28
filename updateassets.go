package pluginctl

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/mattermost/mattermost/server/public/model"
)

//go:embed assets/.editorconfig assets/.gitattributes assets/.nvmrc assets/Makefile assets/*.yml
//go:embed assets/build/*.mk assets/.github/**/*.yml assets/webapp/.npmrc assets/webapp/*.config.js
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

	pluginCtlConfig, err := ParsePluginCtlConfig(manifest)
	if err != nil {
		return fmt.Errorf("failed to parse pluginctl config: %w", err)
	}

	hasWebapp := HasWebappCode(manifest)
	updatedCount := 0

	config := AssetProcessorConfig{
		pluginPath:      pluginPath,
		hasWebapp:       hasWebapp,
		updatedCount:    &updatedCount,
		pluginCtlConfig: pluginCtlConfig,
		manifest:        manifest,
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

// isPathIgnored checks if a path matches any of the ignore patterns.
func isPathIgnored(relativePath string, ignorePatterns []string) (ignored bool, matchedPattern string) {
	for _, pattern := range ignorePatterns {
		// Direct file or path match
		if matched, err := filepath.Match(pattern, relativePath); err == nil && matched {
			return true, pattern
		}

		// Check if the path starts with the pattern (for directory patterns)
		if strings.HasSuffix(pattern, "/") && strings.HasPrefix(relativePath, pattern) {
			return true, pattern
		}

		// Check if any parent directory matches the pattern
		if matchesParentDirectory(relativePath, pattern) {
			return true, pattern
		}

		// Check if any directory component matches the pattern
		if matchesDirectoryComponent(relativePath, pattern) {
			return true, pattern
		}
	}

	return false, ""
}

// matchesParentDirectory checks if any parent directory matches the pattern.
func matchesParentDirectory(relativePath, pattern string) bool {
	dir := filepath.Dir(relativePath)
	for dir != "." && dir != "/" {
		if matched, err := filepath.Match(pattern, dir); err == nil && matched {
			return true
		}
		// Also check direct string match for directory names
		if filepath.Base(dir) == pattern {
			return true
		}
		dir = filepath.Dir(dir)
	}

	return false
}

// matchesDirectoryComponent checks if any directory component matches the pattern.
func matchesDirectoryComponent(relativePath, pattern string) bool {
	parts := strings.Split(relativePath, "/")
	for _, part := range parts {
		if matched, err := filepath.Match(pattern, part); err == nil && matched {
			return true
		}
	}

	return false
}

type AssetProcessorConfig struct {
	pluginPath      string
	hasWebapp       bool
	updatedCount    *int
	pluginCtlConfig *PluginCtlConfig
	manifest        *model.Manifest
}

// GoModule represents information from go.mod file.
type GoModule struct {
	Name    string
	Version string
}

// TemplateContext holds the data available to templates.
type TemplateContext struct {
	Manifest *model.Manifest
	GoModule *GoModule
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

	// Check if path is ignored by pluginctl config
	if ignored, pattern := isPathIgnored(relativePath, config.pluginCtlConfig.IgnoreAssets); ignored {
		Logger.Info("Skipping asset due to ignore pattern", "path", relativePath, "pattern", pattern)

		return nil
	}

	targetPath := filepath.Join(config.pluginPath, relativePath)

	if d.IsDir() {
		return createDirectory(targetPath)
	}

	return processAssetFile(path, targetPath, relativePath, config)
}

func processAssetFile(embeddedPath, targetPath, relativePath string, config AssetProcessorConfig) error {
	shouldUpdate, err := shouldUpdateFile(embeddedPath, targetPath, config)
	if err != nil {
		return err
	}

	if shouldUpdate {
		err = updateFile(embeddedPath, targetPath, relativePath, config)
		if err != nil {
			return err
		}
		(*config.updatedCount)++
	}

	return nil
}

func createDirectory(targetPath string) error {
	if err := os.MkdirAll(targetPath, directoryPermissions); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
	}

	return nil
}

func shouldUpdateFile(embeddedPath, targetPath string, config AssetProcessorConfig) (bool, error) {
	// Process the template to get the final content
	processedContent, err := processTemplate(embeddedPath, config.manifest, config.pluginPath)
	if err != nil {
		return false, fmt.Errorf("failed to process template %s: %w", embeddedPath, err)
	}

	existingContent, err := os.ReadFile(targetPath)
	if err != nil {
		// File doesn't exist or other error, should update
		return true, nil //nolint:nilerr
	}

	return !bytes.Equal(existingContent, processedContent), nil
}

func updateFile(embeddedPath, targetPath, relativePath string, config AssetProcessorConfig) error {
	// Process the template to get the final content
	processedContent, err := processTemplate(embeddedPath, config.manifest, config.pluginPath)
	if err != nil {
		return fmt.Errorf("failed to process template %s: %w", embeddedPath, err)
	}

	parentDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(parentDir, directoryPermissions); err != nil {
		return fmt.Errorf("failed to create parent directory %s: %w", parentDir, err)
	}

	if err := os.WriteFile(targetPath, processedContent, filePermissions); err != nil {
		return fmt.Errorf("failed to write file %s: %w", targetPath, err)
	}

	Logger.Info("Updated file", "path", relativePath)

	return nil
}

// processTemplate processes a template file with the manifest context.
func processTemplate(embeddedPath string, manifest *model.Manifest, pluginPath string) ([]byte, error) {
	// Read the template content
	templateContent, err := assetsFS.ReadFile(embeddedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded file %s: %w", embeddedPath, err)
	}

	// Parse go.mod file to get module information
	goMod, err := parseGoModule(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse go.mod file: %w", err)
	}

	// Create template context
	context := TemplateContext{
		Manifest: manifest,
		GoModule: goMod,
	}

	// Create and parse the template
	tmpl, err := template.New(embeddedPath).Parse(string(templateContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", embeddedPath, err)
	}

	// Execute the template
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, context)
	if err != nil {
		return nil, fmt.Errorf("failed to execute template %s: %w", embeddedPath, err)
	}

	return buf.Bytes(), nil
}

// parseGoModule parses the go.mod file to extract module information.
func parseGoModule(pluginPath string) (*GoModule, error) {
	goModPath := filepath.Join(pluginPath, "go.mod")

	content, err := os.ReadFile(goModPath)
	if err != nil {
		// If go.mod doesn't exist, return nil without error
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to read go.mod file: %w", err)
	}

	goMod := &GoModule{}

	// Using strings.SplitAfter for more efficiency
	for line := range strings.SplitSeq(string(content), "\n") {
		line = strings.TrimSpace(line)

		// Parse module line
		if remainder, found := strings.CutPrefix(line, "module "); found {
			goMod.Name = strings.TrimSpace(remainder)
		}

		// Parse go version line
		if remainder, found := strings.CutPrefix(line, "go "); found {
			goMod.Version = strings.TrimSpace(remainder)

			// We don't need to parse any further
			break
		}
	}

	return goMod, nil
}
