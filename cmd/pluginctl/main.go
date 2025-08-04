package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mattermost/pluginctl"
)

const (
	ExitSuccess   = 0
	ExitError     = 1
	EnvPluginPath = "PLUGINCTL_PLUGIN_PATH"
)

func main() {
	// Initialize logger
	pluginctl.InitLogger()

	var pluginPath string
	var showHelp bool

	flag.StringVar(&pluginPath, "plugin-path", "", "Path to plugin directory (overrides PLUGINCTL_PLUGIN_PATH)")
	flag.BoolVar(&showHelp, "help", false, "Show help information")
	flag.Parse()

	// Show help if requested
	if showHelp {
		showUsage()

		return
	}

	args := flag.Args()
	if len(args) == 0 {
		pluginctl.Logger.Error("No command specified")
		showUsage()
		os.Exit(ExitError)
	}

	command := args[0]
	commandArgs := args[1:]

	// Determine plugin path from flag, environment variable, or current directory
	effectivePluginPath := pluginctl.GetEffectivePluginPath(pluginPath)

	// Validate and update version before running command (except for version command)
	if command != pluginctl.VersionCommand {
		if err := pluginctl.ValidateAndUpdateVersion(effectivePluginPath); err != nil {
			pluginctl.Logger.Error("Version validation failed", "error", err)
			os.Exit(ExitError)
		}
	}

	if err := runCommand(command, commandArgs, effectivePluginPath); err != nil {
		pluginctl.Logger.Error("Command failed", "error", err)
		os.Exit(ExitError)
	}
}

func runCommand(command string, args []string, pluginPath string) error {
	switch command {
	case "info":
		return runInfoCommand(args, pluginPath)
	case "enable":
		return runEnableCommand(args, pluginPath)
	case "disable":
		return runDisableCommand(args, pluginPath)
	case "reset":
		return runResetCommand(args, pluginPath)
	case "deploy":
		return runDeployCommand(args, pluginPath)
	case "updateassets":
		return runUpdateAssetsCommand(args, pluginPath)
	case "manifest":
		return runManifestCommand(args, pluginPath)
	case "logs":
		return runLogsCommand(args, pluginPath)
	case "version":
		return runVersionCommand(args)
	case "create-plugin":
		return runCreatePluginCommand(args, pluginPath)
	case "tools":
		return runToolsCommand(args, pluginPath)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func runInfoCommand(args []string, pluginPath string) error {
	return pluginctl.RunInfoCommand(args, pluginPath)
}

func runVersionCommand(_ []string) error {
	version := pluginctl.GetVersion()
	pluginctl.Logger.Info("pluginctl version", "version", version)

	return nil
}
func runEnableCommand(args []string, pluginPath string) error {
	return pluginctl.RunEnableCommand(args, pluginPath)
}

func runDisableCommand(args []string, pluginPath string) error {
	return pluginctl.RunDisableCommand(args, pluginPath)
}

func runResetCommand(args []string, pluginPath string) error {
	return pluginctl.RunResetCommand(args, pluginPath)
}

func runUpdateAssetsCommand(args []string, pluginPath string) error {
	return pluginctl.RunUpdateAssetsCommand(args, pluginPath)
}

func runManifestCommand(args []string, pluginPath string) error {
	return pluginctl.RunManifestCommand(args, pluginPath)
}

func runLogsCommand(args []string, pluginPath string) error {
	return pluginctl.RunLogsCommand(args, pluginPath)
}

func runDeployCommand(args []string, pluginPath string) error {
	return pluginctl.RunDeployCommand(args, pluginPath)
}

func runCreatePluginCommand(args []string, pluginPath string) error {
	return pluginctl.RunCreatePluginCommand(args, pluginPath)
}

func runToolsCommand(args []string, pluginPath string) error {
	return pluginctl.RunToolsCommand(args, pluginPath)
}

func showUsage() {
	usageText := `pluginctl - Mattermost Plugin Development CLI

Usage:
  pluginctl [global options] <command> [command options] [arguments...]

Global Options:
  --plugin-path PATH   Path to plugin directory (overrides PLUGINCTL_PLUGIN_PATH)
  --help               Show this help message

Commands:
  info           Display plugin information
  enable         Enable plugin in Mattermost server
  disable        Disable plugin in Mattermost server
  reset          Reset plugin (disable then enable)
  deploy         Upload and enable plugin bundle
  updateassets   Update plugin files from embedded assets
  manifest       Manage plugin manifest files
  logs           View plugin logs
  create-plugin  Create a new plugin from template
  tools          Manage development tools (install golangci-lint, gotestsum)
  version        Show version information

Environment Variables:
  PLUGINCTL_PLUGIN_PATH        Default plugin directory path
  MM_LOCALSOCKETPATH           Path to Mattermost local socket
  MM_SERVICESETTINGS_SITEURL   Mattermost server URL
  MM_ADMIN_TOKEN               Admin token for authentication
  MM_ADMIN_USERNAME            Admin username for authentication
  MM_ADMIN_PASSWORD            Admin password for authentication

Use 'pluginctl <command> --help' for detailed information about a command.

For more information about Mattermost plugin development, visit:
https://developers.mattermost.com/integrate/plugins/
`
	pluginctl.Logger.Info(usageText)
}
