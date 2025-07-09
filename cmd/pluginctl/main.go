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

	flag.StringVar(&pluginPath, "plugin-path", "", "Path to plugin directory (overrides PLUGINCTL_PLUGIN_PATH)")
	flag.Parse()

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
	case "updateassets":
		return runUpdateAssetsCommand(args, pluginPath)
	case "help":
		showUsage()

		return nil
	case "version":
		return runVersionCommand(args)
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

func showUsage() {
	usageText := `pluginctl - Mattermost Plugin Development CLI

Usage:
  pluginctl [global options] <command> [command options] [arguments...]

Global Options:
  --plugin-path PATH   Path to plugin directory (overrides PLUGINCTL_PLUGIN_PATH)

Commands:
  info           Display plugin information
  enable         Enable plugin from current directory in Mattermost server
  disable        Disable plugin from current directory in Mattermost server
  reset          Reset plugin from current directory (disable then enable)
  updateassets   Update plugin files from embedded assets
  help           Show this help message
  version        Show version information

Examples:
  pluginctl info                              # Show info for plugin in current directory
  pluginctl --plugin-path /path/to/plugin info # Show info for plugin at specific path
  pluginctl enable                           # Enable plugin from current directory
  pluginctl disable                          # Disable plugin from current directory
  pluginctl reset                            # Reset plugin from current directory (disable then enable)
  pluginctl updateassets                     # Update plugin files from embedded assets
  export PLUGINCTL_PLUGIN_PATH=/path/to/plugin
  pluginctl info                              # Show info using environment variable
  pluginctl version                           # Show version information

Environment Variables:
  PLUGINCTL_PLUGIN_PATH        Default plugin directory path
  MM_LOCALSOCKETPATH           Path to Mattermost local socket
  MM_SERVICESETTINGS_SITEURL   Mattermost server URL
  MM_ADMIN_TOKEN               Admin token for authentication
  MM_ADMIN_USERNAME            Admin username for authentication
  MM_ADMIN_PASSWORD            Admin password for authentication

For more information about Mattermost plugin development, visit:
https://developers.mattermost.com/integrate/plugins/
`
	pluginctl.Logger.Info(usageText)
}
