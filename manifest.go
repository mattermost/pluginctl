package pluginctl

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"
)

// RunManifestCommand implements the 'manifest' command functionality with subcommands.
func RunManifestCommand(args []string, pluginPath string) error {
	if len(args) == 0 {
		return fmt.Errorf("manifest command requires a subcommand: get {{.field_name}}, check")
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
	case "get":
		if len(args) < 2 {
			return fmt.Errorf("get subcommand requires a template expression (e.g., {{.id}}, {{.version}})")
		}
		templateStr := args[1]

		// Parse and execute template with manifest as context
		tmpl, err := template.New("manifest").Parse(templateStr)
		if err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, manifest); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

		fmt.Print(buf.String())
	case "check":
		if err := manifest.IsValid(); err != nil {
			Logger.Error("Plugin manifest validation failed", "error", err)

			return err
		}
		Logger.Info("Plugin manifest is valid")
	default:
		return fmt.Errorf("unknown subcommand: %s. Available subcommands: get {{.field_name}}, check", subcommand)
	}

	return nil
}
