// NOTE: We download tools directly from tarball/binary releases instead of using
// `go get -tool` to prevent modifications to plugin go.mod files on plugins.

package pluginctl

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	defaultGolangciLintVersion = "v2.3.1"
	defaultGotestsumVersion    = "v1.7.0"
	defaultBinDir              = "./build/bin"
	helpFlagLong               = "--help"
	helpFlagShort              = "-h"
	exeExtension               = ".exe"
	tempSuffix                 = "-temp"
	// Platform constants.
	platformDarwin  = "darwin"
	platformWindows = "windows"
	platformLinux   = "linux"
	// Architecture constants.
	archARM64 = "arm64"
	archAMD64 = "amd64"
	arch386   = "386"
	// File permission constants.
	dirPerm  = 0750
	filePerm = 0600
)

// ToolConfig represents configuration for downloading and installing a tool.
type ToolConfig struct {
	Name             string
	Version          string
	GitHubRepo       string
	URLTemplate      string
	FilenameTemplate string
	BinaryPath       string // Path within archive (e.g., "bin/tool" or "tool")
}

var toolConfigs = map[string]ToolConfig{
	"golangci-lint": {
		Name:       "golangci-lint",
		Version:    defaultGolangciLintVersion,
		GitHubRepo: "golangci/golangci-lint",
		URLTemplate: "https://github.com/{repo}/releases/download/{version}/" +
			"golangci-lint-{version_no_v}-{os}-{arch}.tar.gz",
		FilenameTemplate: "golangci-lint-{version_no_v}-{os}-{arch}.tar.gz",
		BinaryPath:       "golangci-lint-{version_no_v}-{os}-{arch}/golangci-lint",
	},
	"gotestsum": {
		Name:       "gotestsum",
		Version:    defaultGotestsumVersion,
		GitHubRepo: "gotestyourself/gotestsum",
		URLTemplate: "https://github.com/{repo}/releases/download/{version}/" +
			"gotestsum_{version_no_v}_{os}_{arch}.tar.gz",
		FilenameTemplate: "gotestsum_{version_no_v}_{os}_{arch}.tar.gz",
		BinaryPath:       "gotestsum",
	},
}

func RunToolsCommand(args []string, pluginPath string) error {
	if len(args) == 0 {
		return showToolsUsage()
	}

	subcommand := args[0]
	subcommandArgs := args[1:]

	switch subcommand {
	case "install":
		return runToolsInstallCommand(subcommandArgs)
	case "help", helpFlagLong, helpFlagShort:
		return showToolsUsage()
	default:
		return fmt.Errorf("unknown tools subcommand: %s", subcommand)
	}
}

func runToolsInstallCommand(args []string) error {
	binDir := defaultBinDir

	for i, arg := range args {
		if arg == helpFlagLong || arg == helpFlagShort {
			return showToolsInstallUsage()
		}
		if arg == "--bin-dir" && i+1 < len(args) {
			binDir = args[i+1]
		}
	}

	Logger.Info("Installing development tools...", "bin-dir", binDir)

	if err := os.MkdirAll(binDir, dirPerm); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	for toolName := range toolConfigs {
		if err := installTool(toolName, binDir); err != nil {
			return fmt.Errorf("failed to install %s: %w", toolName, err)
		}
	}

	Logger.Info("All development tools installed successfully")

	return nil
}

// getPlatform returns the platform string for tool downloads.
func getPlatform() string {
	switch runtime.GOOS {
	case platformDarwin:
		return platformDarwin
	case platformWindows:
		return platformWindows
	default:
		return platformLinux
	}
}

// getArchitecture returns the architecture string for tool downloads.
func getArchitecture() string {
	switch runtime.GOARCH {
	case archARM64:
		return archARM64
	case arch386:
		return arch386
	default:
		return archAMD64
	}
}

// expandTemplate replaces placeholders in template strings.
func expandTemplate(template string, config *ToolConfig, platform, arch string) string {
	versionNoV := strings.TrimPrefix(config.Version, "v")

	replacements := map[string]string{
		"{repo}":         config.GitHubRepo,
		"{version}":      config.Version,
		"{version_no_v}": versionNoV,
		"{os}":           platform,
		"{arch}":         arch,
	}

	result := template
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// downloadAndExtractTool downloads and extracts a tool from GitHub releases.
func downloadAndExtractTool(config *ToolConfig, binDir string) error {
	platform := getPlatform()
	arch := getArchitecture()

	downloadURL := expandTemplate(config.URLTemplate, config, platform, arch)
	binaryPathInArchive := expandTemplate(config.BinaryPath, config, platform, arch)

	Logger.Info("Downloading tool", "tool", config.Name, "url", downloadURL)

	resp, err := downloadToolFromURL(downloadURL, config.Name) //nolint:gosec // Trusted source
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			Logger.Error("Failed to close response body", "error", closeErr)
		}
	}()

	return extractToolFromArchive(resp.Body, config, binDir, binaryPathInArchive)
}

// downloadToolFromURL downloads a tool from the specified URL.
func downloadToolFromURL(downloadURL, toolName string) (*http.Response, error) {
	resp, err := http.Get(downloadURL) //nolint:gosec,noctx // URL from trusted configuration
	if err != nil {
		return nil, fmt.Errorf("failed to download %s: %w", toolName, err)
	}

	if resp.StatusCode != http.StatusOK {
		if closeErr := resp.Body.Close(); closeErr != nil {
			Logger.Error("Failed to close response body", "error", closeErr)
		}

		return nil, fmt.Errorf("failed to download %s: HTTP %d", toolName, resp.StatusCode)
	}

	return resp, nil
}

// extractToolFromArchive extracts the tool binary from a tar.gz archive.
func extractToolFromArchive(reader io.Reader, config *ToolConfig, binDir, binaryPathInArchive string) error {
	gzr, err := gzip.NewReader(reader)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader for %s: %w", config.Name, err)
	}
	defer func() {
		if closeErr := gzr.Close(); closeErr != nil {
			Logger.Error("Failed to close gzip reader", "error", closeErr)
		}
	}()

	tr := tar.NewReader(gzr)

	// Create final binary path
	binaryName := config.Name
	finalBinaryPath := filepath.Join(binDir, fmt.Sprintf("%s-%s", config.Name, config.Version))
	if runtime.GOOS == platformWindows {
		binaryName += exeExtension
		finalBinaryPath += exeExtension
	}

	paths := binaryPaths{
		pathInArchive: binaryPathInArchive,
		binaryName:    binaryName,
		finalPath:     finalBinaryPath,
	}

	return extractBinaryFromTar(tr, config, binDir, paths)
}

// binaryPaths holds path information for binary extraction.
type binaryPaths struct {
	pathInArchive string
	binaryName    string
	finalPath     string
}

// extractBinaryFromTar searches and extracts the binary from a tar archive.
func extractBinaryFromTar(tr *tar.Reader, config *ToolConfig, binDir string, paths binaryPaths) error {
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar archive for %s: %w", config.Name, err)
		}

		if isBinaryFile(header.Name, paths.pathInArchive, paths.binaryName, config.Name) {
			return saveBinaryFile(tr, config, binDir, paths.finalPath)
		}
	}

	return fmt.Errorf("%s binary not found in archive", config.Name)
}

// isBinaryFile checks if the file matches the binary we're looking for.
func isBinaryFile(fileName, binaryPathInArchive, binaryName, configName string) bool {
	return fileName == binaryPathInArchive ||
		strings.HasSuffix(fileName, "/"+binaryName) ||
		strings.HasSuffix(fileName, configName)
}

// saveBinaryFile saves the binary from tar reader to disk.
func saveBinaryFile(tr *tar.Reader, config *ToolConfig, binDir, finalBinaryPath string) error {
	tempPath := filepath.Join(binDir, config.Name+tempSuffix)
	if runtime.GOOS == platformWindows {
		tempPath += exeExtension
	}

	file, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, filePerm)
	if err != nil {
		return fmt.Errorf("failed to create temporary binary file for %s: %w", config.Name, err)
	}

	_, err = io.Copy(file, tr) //nolint:gosec // Archive from trusted source
	if closeErr := file.Close(); closeErr != nil {
		Logger.Error("Failed to close temp file", "error", closeErr)
	}
	if err != nil {
		return fmt.Errorf("failed to write binary file for %s: %w", config.Name, err)
	}

	if err := os.Rename(tempPath, finalBinaryPath); err != nil {
		return fmt.Errorf("failed to rename binary to final path for %s: %w", config.Name, err)
	}

	Logger.Info("Tool installed successfully", "tool", config.Name, "path", finalBinaryPath)

	return nil
}

// installTool installs a single tool by name using its configuration.
func installTool(toolName, binDir string) error {
	config, exists := toolConfigs[toolName]
	if !exists {
		return fmt.Errorf("unknown tool: %s", toolName)
	}

	binaryPath := filepath.Join(binDir, fmt.Sprintf("%s-%s", config.Name, config.Version))
	symlinkPath := filepath.Join(binDir, config.Name)

	if runtime.GOOS == platformWindows {
		binaryPath += exeExtension
		symlinkPath += exeExtension
	}

	if fileExists(binaryPath) {
		return createSymlink(binaryPath, symlinkPath)
	}

	Logger.Info("Installing tool", "tool", config.Name, "version", config.Version)

	if err := downloadAndExtractTool(&config, binDir); err != nil {
		return err
	}

	return createSymlink(binaryPath, symlinkPath)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

func createSymlink(target, link string) error {
	if fileExists(link) {
		if err := os.Remove(link); err != nil {
			return fmt.Errorf("failed to remove existing symlink: %w", err)
		}
	}

	targetRel, err := filepath.Rel(filepath.Dir(link), target)
	if err != nil {
		return fmt.Errorf("failed to calculate relative path: %w", err)
	}

	if err := os.Symlink(targetRel, link); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	Logger.Info("Created symlink", "target", target, "link", link)

	return nil
}

func showToolsUsage() error {
	usageText := `Tools command - Manage development tools

Usage:
  pluginctl tools <subcommand> [options]

Subcommands:
  install    Install development tools (golangci-lint, gotestsum)

Use 'pluginctl tools <subcommand> --help' for detailed information about a subcommand.
`
	Logger.Info(usageText)

	return nil
}

func showToolsInstallUsage() error {
	usageText := `Install development tools

Usage:
  pluginctl tools install

Description:
  Installs the following development tools to ./bin/ directory:
  - golangci-lint ` + defaultGolangciLintVersion + `
  - gotestsum ` + defaultGotestsumVersion + `

  Tools are downloaded with version-specific names (e.g., golangci-lint-v2.3.1)
  to allow version tracking and prevent unnecessary re-downloads.

Options:
  --help, -h    Show this help message
`
	Logger.Info(usageText)

	return nil
}
