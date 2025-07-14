package pluginctl

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
)

func TestLoadPluginManifestFromPath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	tests := []struct {
		name          string
		setupFunc     func(string) error
		path          string
		expectError   bool
		expectedID    string
		expectedName  string
		errorContains string
	}{
		{
			name: "Valid plugin.json",
			setupFunc: func(dir string) error {
				manifest := map[string]interface{}{
					"id":      "com.example.test",
					"name":    "Test Plugin",
					"version": "1.0.0",
				}
				data, _ := json.Marshal(manifest)
				return os.WriteFile(filepath.Join(dir, "plugin.json"), data, 0644)
			},
			path:         tempDir,
			expectError:  false,
			expectedID:   "com.example.test",
			expectedName: "Test Plugin",
		},
		{
			name: "Missing plugin.json",
			setupFunc: func(dir string) error {
				return nil // Don't create any file
			},
			path:          tempDir + "_missing",
			expectError:   true,
			errorContains: "plugin.json not found",
		},
		{
			name: "Invalid JSON",
			setupFunc: func(dir string) error {
				invalidJSON := `{"id": "test", "name": "Test", invalid}`
				return os.WriteFile(filepath.Join(dir, "plugin.json"), []byte(invalidJSON), 0644)
			},
			path:          tempDir,
			expectError:   true,
			errorContains: "failed to parse plugin.json",
		},
		{
			name: "Complex plugin with all fields",
			setupFunc: func(dir string) error {
				manifest := map[string]interface{}{
					"id":                 "com.example.complex",
					"name":               "Complex Plugin",
					"version":            "2.1.0",
					"min_server_version": "7.0.0",
					"description":        "A complex test plugin",
					"server": map[string]interface{}{
						"executables": map[string]string{
							"linux-amd64":   "server/dist/plugin-linux-amd64",
							"darwin-amd64":  "server/dist/plugin-darwin-amd64",
							"windows-amd64": "server/dist/plugin-windows-amd64.exe",
						},
					},
					"webapp": map[string]interface{}{
						"bundle_path": "webapp/dist/main.js",
					},
					"settings_schema": map[string]interface{}{
						"header": "Complex Plugin Settings",
						"settings": []map[string]interface{}{
							{
								"key":          "enable_feature",
								"display_name": "Enable Feature",
								"type":         "bool",
								"default":      true,
							},
						},
					},
				}
				data, _ := json.Marshal(manifest)
				return os.WriteFile(filepath.Join(dir, "plugin.json"), data, 0644)
			},
			path:         tempDir,
			expectError:  false,
			expectedID:   "com.example.complex",
			expectedName: "Complex Plugin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test data
			if err := tt.setupFunc(tt.path); err != nil {
				t.Fatalf("Failed to setup test: %v", err)
			}

			// Test the function
			manifest, err := LoadPluginManifestFromPath(tt.path)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, but got: %v", tt.errorContains, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if manifest == nil {
				t.Fatal("Expected manifest but got nil")
			}

			if manifest.Id != tt.expectedID {
				t.Errorf("Expected ID %q, got %q", tt.expectedID, manifest.Id)
			}

			if manifest.Name != tt.expectedName {
				t.Errorf("Expected Name %q, got %q", tt.expectedName, manifest.Name)
			}
		})
	}
}

func TestLoadPluginManifest(t *testing.T) {
	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Create temporary directory and change to it
	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Test with no plugin.json
	_, err = LoadPluginManifest()
	if err == nil {
		t.Error("Expected error when no plugin.json exists")
	}

	// Create a valid plugin.json
	manifest := map[string]interface{}{
		"id":      "com.example.current",
		"name":    "Current Dir Plugin",
		"version": "1.0.0",
	}
	data, _ := json.Marshal(manifest)
	if err := os.WriteFile("plugin.json", data, 0644); err != nil {
		t.Fatalf("Failed to create plugin.json: %v", err)
	}

	// Test with valid plugin.json
	result, err := LoadPluginManifest()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Id != "com.example.current" {
		t.Errorf("Expected ID 'com.example.current', got %q", result.Id)
	}
}

func TestGetEffectivePluginPath(t *testing.T) {
	// Save original environment
	originalEnv := os.Getenv("PLUGINCTL_PLUGIN_PATH")
	defer os.Setenv("PLUGINCTL_PLUGIN_PATH", originalEnv)

	// Save original working directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	tests := []struct {
		name        string
		flagPath    string
		envPath     string
		expectedDir string
	}{
		{
			name:        "Flag path takes priority",
			flagPath:    "/path/from/flag",
			envPath:     "/path/from/env",
			expectedDir: "/path/from/flag",
		},
		{
			name:        "Environment variable when no flag",
			flagPath:    "",
			envPath:     "/path/from/env",
			expectedDir: "/path/from/env",
		},
		{
			name:        "Current directory when no flag or env",
			flagPath:    "",
			envPath:     "",
			expectedDir: originalDir,
		},
		{
			name:        "Empty flag falls back to env",
			flagPath:    "",
			envPath:     "/path/from/env",
			expectedDir: "/path/from/env",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			os.Setenv("PLUGINCTL_PLUGIN_PATH", tt.envPath)

			result := GetEffectivePluginPath(tt.flagPath)

			if result != tt.expectedDir {
				t.Errorf("Expected path %q, got %q", tt.expectedDir, result)
			}
		})
	}
}

func TestPluginManifestValidation(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name         string
		manifestData map[string]interface{}
		expectValid  bool
	}{
		{
			name: "Minimal valid manifest",
			manifestData: map[string]interface{}{
				"id":      "com.example.minimal",
				"name":    "Minimal Plugin",
				"version": "1.0.0",
			},
			expectValid: true,
		},
		{
			name: "Manifest with server executables",
			manifestData: map[string]interface{}{
				"id":      "com.example.server",
				"name":    "Server Plugin",
				"version": "1.0.0",
				"server": map[string]interface{}{
					"executables": map[string]string{
						"linux-amd64": "server/plugin",
					},
				},
			},
			expectValid: true,
		},
		{
			name: "Manifest with webapp bundle",
			manifestData: map[string]interface{}{
				"id":      "com.example.webapp",
				"name":    "Webapp Plugin",
				"version": "1.0.0",
				"webapp": map[string]interface{}{
					"bundle_path": "webapp/dist/main.js",
				},
			},
			expectValid: true,
		},
		{
			name: "Manifest with settings schema",
			manifestData: map[string]interface{}{
				"id":      "com.example.settings",
				"name":    "Settings Plugin",
				"version": "1.0.0",
				"settings_schema": map[string]interface{}{
					"header": "Plugin Settings",
					"settings": []map[string]interface{}{
						{
							"key":     "test_setting",
							"type":    "text",
							"default": "value",
						},
					},
				},
			},
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create plugin.json file
			data, err := json.Marshal(tt.manifestData)
			if err != nil {
				t.Fatalf("Failed to marshal test data: %v", err)
			}

			pluginPath := filepath.Join(tempDir, "plugin.json")
			if err := os.WriteFile(pluginPath, data, 0644); err != nil {
				t.Fatalf("Failed to write plugin.json: %v", err)
			}

			// Load and validate manifest
			manifest, err := LoadPluginManifestFromPath(tempDir)

			if tt.expectValid {
				if err != nil {
					t.Errorf("Expected valid manifest but got error: %v", err)
				}
				if manifest == nil {
					t.Error("Expected manifest but got nil")
				}
			} else {
				if err == nil {
					t.Error("Expected error for invalid manifest but got nil")
				}
			}

			// Clean up for next test
			os.Remove(pluginPath)
		})
	}
}

func TestHasServerCodeAndWebappCode(t *testing.T) {
	tests := []struct {
		name           string
		manifest       *model.Manifest
		expectedServer bool
		expectedWebapp bool
	}{
		{
			name: "Plugin with both server and webapp",
			manifest: &model.Manifest{
				Server: &model.ManifestServer{
					Executables: map[string]string{
						"linux-amd64": "server/plugin",
					},
				},
				Webapp: &model.ManifestWebapp{
					BundlePath: "webapp/dist/main.js",
				},
			},
			expectedServer: true,
			expectedWebapp: true,
		},
		{
			name: "Plugin with server only",
			manifest: &model.Manifest{
				Server: &model.ManifestServer{
					Executables: map[string]string{
						"linux-amd64": "server/plugin",
					},
				},
			},
			expectedServer: true,
			expectedWebapp: false,
		},
		{
			name: "Plugin with webapp only",
			manifest: &model.Manifest{
				Webapp: &model.ManifestWebapp{
					BundlePath: "webapp/dist/main.js",
				},
			},
			expectedServer: false,
			expectedWebapp: true,
		},
		{
			name:           "Plugin with neither",
			manifest:       &model.Manifest{},
			expectedServer: false,
			expectedWebapp: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverResult := HasServerCode(tt.manifest)
			webappResult := HasWebappCode(tt.manifest)

			if serverResult != tt.expectedServer {
				t.Errorf("hasServerCode() = %v, expected %v", serverResult, tt.expectedServer)
			}

			if webappResult != tt.expectedWebapp {
				t.Errorf("hasWebappCode() = %v, expected %v", webappResult, tt.expectedWebapp)
			}
		})
	}
}

func TestParsePluginCtlConfig(t *testing.T) {
	tests := []struct {
		name           string
		manifest       *model.Manifest
		expectedConfig *PluginCtlConfig
		expectError    bool
		errorContains  string
	}{
		{
			name: "Manifest with valid pluginctl config",
			manifest: &model.Manifest{
				Props: map[string]interface{}{
					"pluginctl": map[string]interface{}{
						"ignore_assets": []string{"*.test.js", "build/", "temp/**"},
					},
				},
			},
			expectedConfig: &PluginCtlConfig{
				IgnoreAssets: []string{"*.test.js", "build/", "temp/**"},
			},
			expectError: false,
		},
		{
			name: "Manifest with no props",
			manifest: &model.Manifest{
				Props: nil,
			},
			expectedConfig: &PluginCtlConfig{
				IgnoreAssets: []string{},
			},
			expectError: false,
		},
		{
			name: "Manifest with empty props",
			manifest: &model.Manifest{
				Props: map[string]interface{}{},
			},
			expectedConfig: &PluginCtlConfig{
				IgnoreAssets: []string{},
			},
			expectError: false,
		},
		{
			name: "Manifest with no pluginctl config",
			manifest: &model.Manifest{
				Props: map[string]interface{}{
					"other": "value",
				},
			},
			expectedConfig: &PluginCtlConfig{
				IgnoreAssets: []string{},
			},
			expectError: false,
		},
		{
			name: "Manifest with empty pluginctl config",
			manifest: &model.Manifest{
				Props: map[string]interface{}{
					"pluginctl": map[string]interface{}{},
				},
			},
			expectedConfig: &PluginCtlConfig{
				IgnoreAssets: []string{},
			},
			expectError: false,
		},
		{
			name: "Manifest with invalid pluginctl config",
			manifest: &model.Manifest{
				Props: map[string]interface{}{
					"pluginctl": "invalid",
				},
			},
			expectedConfig: nil,
			expectError:    true,
			errorContains:  "failed to parse pluginctl config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParsePluginCtlConfig(tt.manifest)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q but got: %v", tt.errorContains, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if config == nil {
				t.Error("Expected config but got nil")
				return
			}

			if len(config.IgnoreAssets) != len(tt.expectedConfig.IgnoreAssets) {
				t.Errorf("Expected %d ignore assets but got %d", len(tt.expectedConfig.IgnoreAssets), len(config.IgnoreAssets))
				return
			}

			for i, expected := range tt.expectedConfig.IgnoreAssets {
				if config.IgnoreAssets[i] != expected {
					t.Errorf("Expected ignore asset %d to be %q but got %q", i, expected, config.IgnoreAssets[i])
				}
			}
		})
	}
}

func TestIsPathIgnored(t *testing.T) {
	tests := []struct {
		name           string
		relativePath   string
		ignorePatterns []string
		expectedIgnore bool
		expectedPattern string
	}{
		{
			name:           "No ignore patterns",
			relativePath:   "webapp/dist/main.js",
			ignorePatterns: []string{},
			expectedIgnore: false,
			expectedPattern: "",
		},
		{
			name:           "Direct file match",
			relativePath:   "test.js",
			ignorePatterns: []string{"*.js"},
			expectedIgnore: true,
			expectedPattern: "*.js",
		},
		{
			name:           "Directory pattern with slash",
			relativePath:   "build/output.js",
			ignorePatterns: []string{"build/"},
			expectedIgnore: true,
			expectedPattern: "build/",
		},
		{
			name:           "Directory pattern without slash",
			relativePath:   "build/output.js",
			ignorePatterns: []string{"build"},
			expectedIgnore: true,
			expectedPattern: "build",
		},
		{
			name:           "Nested directory match",
			relativePath:   "webapp/dist/main.js",
			ignorePatterns: []string{"dist"},
			expectedIgnore: true,
			expectedPattern: "dist",
		},
		{
			name:           "Multiple patterns - first match",
			relativePath:   "test.js",
			ignorePatterns: []string{"*.js", "*.css"},
			expectedIgnore: true,
			expectedPattern: "*.js",
		},
		{
			name:           "Multiple patterns - second match",
			relativePath:   "style.css",
			ignorePatterns: []string{"*.js", "*.css"},
			expectedIgnore: true,
			expectedPattern: "*.css",
		},
		{
			name:           "No match",
			relativePath:   "README.md",
			ignorePatterns: []string{"*.js", "*.css"},
			expectedIgnore: false,
			expectedPattern: "",
		},
		{
			name:           "Complex path with match",
			relativePath:   "webapp/node_modules/package/file.js",
			ignorePatterns: []string{"node_modules"},
			expectedIgnore: true,
			expectedPattern: "node_modules",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ignored, pattern := isPathIgnored(tt.relativePath, tt.ignorePatterns)

			if ignored != tt.expectedIgnore {
				t.Errorf("Expected ignore result %v but got %v", tt.expectedIgnore, ignored)
			}

			if pattern != tt.expectedPattern {
				t.Errorf("Expected pattern %q but got %q", tt.expectedPattern, pattern)
			}
		})
	}
}
