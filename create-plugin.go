package pluginctl

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	starterTemplateURL    = "https://github.com/mattermost/mattermost-plugin-starter-template.git"
	starterTemplateWebURL = "https://github.com/mattermost/mattermost-plugin-starter-template"
	pluginPrefix          = "mattermost-plugin-"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Margin(1, 0)

	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("36")).
			Bold(true)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Background(lipgloss.Color("235")).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39"))
)

type inputModel struct {
	input       string
	cursor      int
	placeholder string
	prompt      string
	validation  func(string) string
	submitted   bool
	fixedPrefix string
}

func (m *inputModel) Init() tea.Cmd {
	return nil
}

func (m *inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	return m.handleKeyMsg(keyMsg)
}

func (m *inputModel) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	//nolint:exhaustive
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEnter:
		if validationErr := m.validation(m.input); validationErr == "" {
			m.submitted = true

			return m, tea.Quit
		}
	case tea.KeyLeft:
		m.handleCursorLeft()
	case tea.KeyRight:
		m.handleCursorRight()
	case tea.KeyBackspace:
		m.handleBackspace()
	case tea.KeyDelete:
		m.handleDelete()
	case tea.KeyHome:
		m.cursor = 0
	case tea.KeyEnd:
		m.cursor = len(m.input)
	case tea.KeyRunes:
		m.handleRunes(msg.Runes)
	}

	return m, nil
}

func (m *inputModel) handleCursorLeft() {
	if m.cursor > 0 {
		m.cursor--
	}
}

func (m *inputModel) handleCursorRight() {
	if m.cursor < len(m.input) {
		m.cursor++
	}
}

func (m *inputModel) handleBackspace() {
	if m.cursor > 0 {
		m.input = m.input[:m.cursor-1] + m.input[m.cursor:]
		m.cursor--
	}
}

func (m *inputModel) handleDelete() {
	if m.cursor < len(m.input) {
		m.input = m.input[:m.cursor] + m.input[m.cursor+1:]
	}
}

func (m *inputModel) handleRunes(runes []rune) {
	m.input = m.input[:m.cursor] + string(runes) + m.input[m.cursor:]
	m.cursor += len(runes)
}

func (m *inputModel) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Create Mattermost Plugin"))
	s.WriteString("\n\n")

	s.WriteString(promptStyle.Render(m.prompt))
	s.WriteString("\n")

	// Show input field with fixed prefix
	displayValue := m.fixedPrefix + m.input
	if m.input == "" {
		displayValue = m.fixedPrefix + m.placeholder
	}

	// Add cursor (adjust position for fixed prefix)
	cursorPos := len(m.fixedPrefix) + m.cursor
	if cursorPos <= len(displayValue) {
		left := displayValue[:cursorPos]
		right := displayValue[cursorPos:]
		if cursorPos == len(displayValue) {
			displayValue = left + "█"
		} else {
			displayValue = left + "█" + right[1:]
		}
	}

	s.WriteString(inputStyle.Render(displayValue))
	s.WriteString("\n")

	// Show validation message
	if m.input != "" {
		if validationErr := m.validation(m.input); validationErr != "" {
			s.WriteString(errorStyle.Render("✗ " + validationErr))
		} else {
			fullName := m.fixedPrefix + m.input
			s.WriteString(successStyle.Render("✓ Valid plugin name: " + fullName))
		}
		s.WriteString("\n")
	}

	s.WriteString(infoStyle.Render("\nPress Enter to continue, Ctrl+C to cancel"))

	return s.String()
}

func validatePluginSuffix(suffix string) string {
	if suffix == "" {
		return "Plugin name cannot be empty"
	}

	// Check for valid Go module name (simplified)
	validName := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !validName.MatchString(suffix) {
		return "Plugin name contains invalid characters"
	}

	// Check for reasonable length
	if len(suffix) < 2 {
		return "Plugin name must be at least 2 characters long"
	}

	return ""
}

func validateModuleName(name string) string {
	if name == "" {
		return "Module name cannot be empty"
	}

	// Check for valid Go module name
	validModule := regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`)
	if !validModule.MatchString(name) {
		return "Module name contains invalid characters"
	}

	// Should look like a valid repository path
	if !strings.Contains(name, "/") {
		return "Module name should be a repository path (e.g., github.com/user/repo)"
	}

	return ""
}

func toTitleCase(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if word != "" {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, " ")
}

func promptForPluginName() (string, error) {
	model := &inputModel{
		prompt:      "Enter plugin name:",
		placeholder: "example",
		validation:  validatePluginSuffix,
		fixedPrefix: pluginPrefix,
	}

	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run input prompt: %w", err)
	}

	result, ok := finalModel.(*inputModel)
	if !ok {
		return "", fmt.Errorf("unexpected model type")
	}
	if !result.submitted {
		return "", fmt.Errorf("operation canceled")
	}

	// Return the full plugin name (prefix + suffix)
	return result.fixedPrefix + result.input, nil
}

func promptForModuleName(pluginName string) (string, error) {
	// Generate a sensible default based on the plugin name
	defaultModule := fmt.Sprintf("github.com/user/%s", pluginName)
	model := &inputModel{
		prompt:      "Enter Go module name (repository path):",
		placeholder: defaultModule,
		validation:  validateModuleName,
	}

	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run input prompt: %w", err)
	}

	result, ok := finalModel.(*inputModel)
	if !ok {
		return "", fmt.Errorf("unexpected model type")
	}
	if !result.submitted {
		return "", fmt.Errorf("operation canceled")
	}

	return result.input, nil
}

func parseCreatePluginFlags(args []string) (pluginName, moduleName string, err error) {
	// Parse flags similar to how logs command handles --watch
	for i, arg := range args {
		switch arg {
		case "--name":
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("--name flag requires a value")
			}
			pluginName = args[i+1]
		case "--module":
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("--module flag requires a value")
			}
			moduleName = args[i+1]
		}
	}

	// Validate and process plugin name if provided
	if pluginName != "" {
		pluginName, err = validateAndProcessPluginName(pluginName)
		if err != nil {
			return "", "", err
		}
	}

	// Validate module name if provided
	if moduleName != "" {
		if validationErr := validateModuleName(moduleName); validationErr != "" {
			return "", "", fmt.Errorf("invalid module name: %s", validationErr)
		}
	}

	return pluginName, moduleName, nil
}

// validateAndProcessPluginName checks if the plugin name has the correct prefix and adds it if necessary.
// Example:
// - If the input is "my-plugin", it returns "mattermost-plugin-my-plugin".
// - If the input is "mattermost-plugin-my-plugin", it returns "mattermost-plugin-my-plugin".
func validateAndProcessPluginName(name string) (string, error) {
	// Check if the name already has the prefix
	if !strings.HasPrefix(name, pluginPrefix) {
		// If not, validate the suffix and add prefix
		if err := validatePluginSuffix(name); err != "" {
			return "", fmt.Errorf("invalid plugin name: %s", err)
		}

		return pluginPrefix + name, nil
	}

	// If it has the prefix, validate the suffix part
	suffix := strings.TrimPrefix(name, pluginPrefix)
	if err := validatePluginSuffix(suffix); err != "" {
		return "", fmt.Errorf("invalid plugin name: %s", err)
	}

	return name, nil
}

func RunCreatePluginCommand(args []string, pluginPath string) error {
	// Parse flags
	var pluginName, moduleName string
	var err error

	pluginName, moduleName, err = parseCreatePluginFlags(args)
	if err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	Logger.Info("Starting plugin creation process")

	// If flags were not provided, fall back to interactive mode
	if pluginName == "" {
		pluginName, err = promptForPluginName()
		if err != nil {
			return fmt.Errorf("failed to get plugin name: %w", err)
		}
	}

	if moduleName == "" {
		moduleName, err = promptForModuleName(pluginName)
		if err != nil {
			return fmt.Errorf("failed to get module name: %w", err)
		}
	}

	Logger.Info("Creating plugin", "name", pluginName, "module", moduleName)

	// Check if directory already exists
	if _, err := os.Stat(pluginName); err == nil {
		return fmt.Errorf("directory '%s' already exists", pluginName)
	}

	// Clone the starter template
	Logger.Info("Cloning starter template repository")
	if err := cloneStarterTemplate(pluginName); err != nil {
		return fmt.Errorf("failed to clone starter template: %w", err)
	}

	// Process the template
	Logger.Info("Processing template files")
	if err := processPluginTemplate(pluginName, moduleName); err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	// Update assets
	Logger.Info("Updating plugin assets")
	if err := RunUpdateAssetsCommand([]string{}, pluginName); err != nil {
		return fmt.Errorf("failed to update assets: %w", err)
	}

	// Run go mod tidy
	Logger.Info("Running go mod tidy")
	if err := runGoModTidy(pluginName); err != nil {
		return fmt.Errorf("failed to run go mod tidy: %w", err)
	}

	// Initialize git and create initial commit
	Logger.Info("Initializing git repository")
	if err := initializeGitRepo(pluginName); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	Logger.Info("Plugin created successfully!", "name", pluginName, "path", pluginName)
	fmt.Printf("\n%s\n", successStyle.Render("✓ Plugin created successfully!"))
	fmt.Printf("%s\n", infoStyle.Render(fmt.Sprintf("Plugin '%s' has been created in directory: %s",
		pluginName, pluginName)))
	fmt.Printf("%s\n", infoStyle.Render("Next steps:"))
	fmt.Printf("%s\n", infoStyle.Render("  1. cd "+pluginName))
	fmt.Printf("%s\n", infoStyle.Render("  2. Review and modify the plugin.json file"))
	fmt.Printf("%s\n", infoStyle.Render("  3. Start developing your plugin!"))

	return nil
}

func cloneStarterTemplate(pluginName string) error {
	cmd := exec.Command("git", "clone", "--depth", "1", starterTemplateURL, pluginName)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	// Remove the .git directory from the template
	gitDir := filepath.Join(pluginName, ".git")
	if err := os.RemoveAll(gitDir); err != nil {
		return fmt.Errorf("failed to remove .git directory: %w", err)
	}

	return nil
}

func processPluginTemplate(pluginName, moduleName string) error {
	// Update go.mod
	if err := updateGoMod(pluginName, moduleName); err != nil {
		return fmt.Errorf("failed to update go.mod: %w", err)
	}

	// Update plugin.json
	if err := updatePluginJSON(pluginName, moduleName); err != nil {
		return fmt.Errorf("failed to update plugin.json: %w", err)
	}

	// Update Go code references
	if err := updateGoCodeReferences(pluginName, moduleName); err != nil {
		return fmt.Errorf("failed to update Go code references: %w", err)
	}

	// Update README.md
	if err := updateReadme(pluginName, moduleName); err != nil {
		return fmt.Errorf("failed to update README.md: %w", err)
	}

	// Update hello.html
	if err := updateHelloHTML(pluginName); err != nil {
		return fmt.Errorf("failed to update hello.html: %w", err)
	}

	return nil
}

func updateGoMod(pluginName, moduleName string) error {
	goModPath := filepath.Join(pluginName, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return fmt.Errorf("failed to read go.mod: %w", err)
	}

	// Replace the module name
	updated := strings.Replace(string(content),
		"github.com/mattermost/mattermost-plugin-starter-template", moduleName, 1)

	if err := os.WriteFile(goModPath, []byte(updated), filePermissions); err != nil {
		return fmt.Errorf("failed to write go.mod: %w", err)
	}

	return nil
}

func updatePluginJSON(pluginName, moduleName string) error {
	pluginJSONPath := filepath.Join(pluginName, "plugin.json")

	// Load the manifest using the existing function
	manifest, err := LoadPluginManifestFromPath(pluginName)
	if err != nil {
		return fmt.Errorf("failed to load plugin manifest: %w", err)
	}

	// Update the plugin ID (remove the prefix for the ID)
	pluginID := strings.TrimPrefix(pluginName, pluginPrefix)
	manifest.Id = "com.mattermost." + pluginID

	// Update display name
	displayName := strings.ReplaceAll(pluginID, "-", " ")
	displayName = toTitleCase(displayName)
	manifest.Name = displayName

	// Update homepage_url and support_url if module is a GitHub repository
	if strings.HasPrefix(moduleName, "github.com/") {
		newURL := "https://" + moduleName
		manifest.HomepageURL = newURL
		manifest.SupportURL = newURL + "/issues"
	}

	// Update icon path
	if manifest.IconPath == "assets/starter-template-icon.svg" {
		manifest.IconPath = "assets/" + pluginID + "-icon.svg"
	}

	// Write the updated manifest back to JSON
	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal plugin manifest: %w", err)
	}

	if err := os.WriteFile(pluginJSONPath, manifestJSON, filePermissions); err != nil {
		return fmt.Errorf("failed to write plugin.json: %w", err)
	}

	return nil
}

func updateGoCodeReferences(pluginName, moduleName string) error {
	// Walk through Go files and update import references
	return filepath.Walk(pluginName, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		// Replace import references
		updated := strings.Replace(string(content),
			"github.com/mattermost/mattermost-plugin-starter-template", moduleName, -1)

		// Update plugin ID in comments (like api.go)
		pluginID := strings.TrimPrefix(pluginName, pluginPrefix)
		updated = strings.Replace(updated, "com.mattermost.plugin-starter-template",
			"com.mattermost."+pluginID, -1)

		if err := os.WriteFile(path, []byte(updated), filePermissions); err != nil {
			return fmt.Errorf("failed to write file %s: %w", path, err)
		}

		return nil
	})
}

func runGoModTidy(pluginName string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = pluginName
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}

	return nil
}

func initializeGitRepo(pluginName string) error {
	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = pluginName
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git init failed: %w", err)
	}

	// Add all files
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = pluginName
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	// Create initial commit
	cmd = exec.Command("git", "commit", "-m", "Initial commit from mattermost-plugin-starter-template")
	cmd.Dir = pluginName
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	return nil
}

func updateReadme(pluginName, moduleName string) error {
	readmePath := filepath.Join(pluginName, "README.md")
	content, err := os.ReadFile(readmePath)
	if err != nil {
		// README.md might not exist, which is fine
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("failed to read README.md: %w", err)
	}

	updated := string(content)

	// Update all GitHub URLs to point to the new repository
	if strings.HasPrefix(moduleName, "github.com/") {
		newURL := "https://" + moduleName
		updated = strings.Replace(updated, starterTemplateWebURL, newURL, -1)
	}

	// Update plugin ID in comments and examples
	pluginID := strings.TrimPrefix(pluginName, pluginPrefix)
	updated = strings.Replace(updated, "com.mattermost.plugin-starter-template",
		"com.mattermost."+pluginID, -1)

	// Update clone command example
	updated = strings.Replace(updated, "com.example.my-plugin", "com.example."+pluginID, -1)

	if err := os.WriteFile(readmePath, []byte(updated), filePermissions); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	return nil
}

func updateHelloHTML(pluginName string) error {
	helloPath := filepath.Join(pluginName, "public", "hello.html")
	content, err := os.ReadFile(helloPath)
	if err != nil {
		// hello.html might not exist, which is fine
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("failed to read hello.html: %w", err)
	}

	pluginID := strings.TrimPrefix(pluginName, pluginPrefix)
	updated := strings.Replace(string(content), "com.mattermost.plugin-starter-template",
		"com.mattermost."+pluginID, -1)

	if err := os.WriteFile(helloPath, []byte(updated), filePermissions); err != nil {
		return fmt.Errorf("failed to write hello.html: %w", err)
	}

	return nil
}
