package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
)

func TestPrintPluginInfo(t *testing.T) {
	tests := []struct {
		name     string
		manifest *model.Manifest
		expected []string
	}{
		{
			name: "Complete plugin with all features",
			manifest: &model.Manifest{
				Id:               "com.example.testplugin",
				Name:             "Test Plugin",
				Version:          "1.0.0",
				MinServerVersion: "7.0.0",
				Description:      "A test plugin for unit testing",
				Server: &model.ManifestServer{
					Executables: map[string]string{
						"linux-amd64":   "server/dist/plugin-linux-amd64",
						"darwin-amd64":  "server/dist/plugin-darwin-amd64",
						"windows-amd64": "server/dist/plugin-windows-amd64.exe",
					},
				},
				Webapp: &model.ManifestWebapp{
					BundlePath: "webapp/dist/main.js",
				},
				SettingsSchema: &model.PluginSettingsSchema{
					Header: "Test Settings",
				},
			},
			expected: []string{
				"Plugin Information:",
				"ID:              com.example.testplugin",
				"Name:            Test Plugin",
				"Version:         1.0.0",
				"Min MM Version:  7.0.0",
				"Description:     A test plugin for unit testing",
				"Code Components:",
				"Server Code:     Yes",
				"linux-amd64",
				"darwin-amd64",
				"windows-amd64",
				"Webapp Code:     Yes",
				"Bundle Path:   webapp/dist/main.js",
				"Settings Schema: Yes",
			},
		},
		{
			name: "Minimal plugin with no optional fields",
			manifest: &model.Manifest{
				Id:      "com.example.minimal",
				Name:    "Minimal Plugin",
				Version: "0.1.0",
			},
			expected: []string{
				"Plugin Information:",
				"ID:              com.example.minimal",
				"Name:            Minimal Plugin",
				"Version:         0.1.0",
				"Min MM Version:  Not specified",
				"Code Components:",
				"Server Code:     No",
				"Webapp Code:     No",
				"Settings Schema: No",
			},
		},
		{
			name: "Plugin with server code only",
			manifest: &model.Manifest{
				Id:               "com.example.serveronly",
				Name:             "Server Only Plugin",
				Version:          "2.0.0",
				MinServerVersion: "8.0.0",
				Server: &model.ManifestServer{
					Executables: map[string]string{
						"linux-amd64": "server/plugin",
					},
				},
			},
			expected: []string{
				"Plugin Information:",
				"ID:              com.example.serveronly",
				"Name:            Server Only Plugin",
				"Version:         2.0.0",
				"Min MM Version:  8.0.0",
				"Server Code:     Yes",
				"linux-amd64",
				"Webapp Code:     No",
				"Settings Schema: No",
			},
		},
		{
			name: "Plugin with webapp code only",
			manifest: &model.Manifest{
				Id:      "com.example.webapponly",
				Name:    "Webapp Only Plugin",
				Version: "1.5.0",
				Webapp: &model.ManifestWebapp{
					BundlePath: "dist/bundle.js",
				},
			},
			expected: []string{
				"Plugin Information:",
				"ID:              com.example.webapponly",
				"Name:            Webapp Only Plugin",
				"Version:         1.5.0",
				"Min MM Version:  Not specified",
				"Server Code:     No",
				"Webapp Code:     Yes",
				"Bundle Path:   dist/bundle.js",
				"Settings Schema: No",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run the function
			err := PrintPluginInfo(tt.manifest)
			if err != nil {
				t.Fatalf("PrintPluginInfo returned error: %v", err)
			}

			// Restore stdout and get output
			w.Close()
			os.Stdout = oldStdout

			output, _ := io.ReadAll(r)
			outputStr := string(output)

			// Check that all expected strings are present
			for _, expected := range tt.expected {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", expected, outputStr)
				}
			}
		})
	}
}

func TestHasServerCode(t *testing.T) {
	tests := []struct {
		name     string
		manifest *model.Manifest
		expected bool
	}{
		{
			name: "Plugin with server executables",
			manifest: &model.Manifest{
				Server: &model.ManifestServer{
					Executables: map[string]string{
						"linux-amd64": "server/plugin",
					},
				},
			},
			expected: true,
		},
		{
			name: "Plugin with empty server executables",
			manifest: &model.Manifest{
				Server: &model.ManifestServer{
					Executables: map[string]string{},
				},
			},
			expected: false,
		},
		{
			name: "Plugin with nil server",
			manifest: &model.Manifest{
				Server: nil,
			},
			expected: false,
		},
		{
			name:     "Plugin with no server field",
			manifest: &model.Manifest{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasServerCode(tt.manifest)
			if result != tt.expected {
				t.Errorf("hasServerCode() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestHasWebappCode(t *testing.T) {
	tests := []struct {
		name     string
		manifest *model.Manifest
		expected bool
	}{
		{
			name: "Plugin with webapp bundle",
			manifest: &model.Manifest{
				Webapp: &model.ManifestWebapp{
					BundlePath: "webapp/dist/main.js",
				},
			},
			expected: true,
		},
		{
			name: "Plugin with empty webapp bundle path",
			manifest: &model.Manifest{
				Webapp: &model.ManifestWebapp{
					BundlePath: "",
				},
			},
			expected: false,
		},
		{
			name: "Plugin with nil webapp",
			manifest: &model.Manifest{
				Webapp: nil,
			},
			expected: false,
		},
		{
			name:     "Plugin with no webapp field",
			manifest: &model.Manifest{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasWebappCode(tt.manifest)
			if result != tt.expected {
				t.Errorf("hasWebappCode() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestInfoCommandWithPath_InvalidPath(t *testing.T) {
	// Test with non-existent directory
	err := InfoCommandWithPath("/non/existent/path")
	if err == nil {
		t.Error("Expected error for non-existent path, but got nil")
	}

	expectedErrMsg := "plugin.json not found"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error to contain %q, but got: %v", expectedErrMsg, err)
	}
}

// captureOutput captures stdout during function execution
func captureOutput(fn func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestPrintPluginInfo_OutputFormat(t *testing.T) {
	manifest := &model.Manifest{
		Id:      "test.plugin",
		Name:    "Test Plugin",
		Version: "1.0.0",
	}

	output := captureOutput(func() {
		PrintPluginInfo(manifest)
	})

	// Check for proper formatting

	// Should have header separators
	if !strings.Contains(output, "==================") {
		t.Error("Output should contain header separators")
	}

	// Should have proper sections
	if !strings.Contains(output, "Plugin Information:") {
		t.Error("Output should contain 'Plugin Information:' header")
	}

	if !strings.Contains(output, "Code Components:") {
		t.Error("Output should contain 'Code Components:' header")
	}

	// Should have proper field formatting
	expectedFields := []string{
		"ID:              test.plugin",
		"Name:            Test Plugin",
		"Version:         1.0.0",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Output should contain field: %q", field)
		}
	}
}
