package pluginctl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
)

const (
	logsPerPage        = 100
	defaultLogsPerPage = 500
	timeStampFormat    = "2006-01-02 15:04:05.000 Z07:00"
)

// RunLogsCommand executes the logs command with optional --watch flag.
func RunLogsCommand(args []string, pluginPath string) error {
	helpText := `View plugin logs

Usage:
  pluginctl logs [options]

Options:
  --watch           Follow logs in real-time
  --help, -h        Show this help message

Description:
  Views plugin logs from the Mattermost server. By default, shows recent log
  entries. Use --watch to follow logs in real-time.

  Note: JSON output for file logs must be enabled in Mattermost configuration
  (LogSettings.FileJson) for this command to work.

Examples:
  pluginctl logs                              # View recent plugin logs
  pluginctl logs --watch                      # Watch plugin logs in real-time
  pluginctl --plugin-path /path/to/plugin logs # View logs for plugin at specific path`

	// Check for help flag
	if CheckForHelpFlag(args, helpText) {
		return nil
	}

	// Check for --watch flag
	watch := false
	if len(args) > 0 && args[0] == "--watch" {
		watch = true
	}

	if watch {
		return runPluginCommand(args, pluginPath, watchPluginLogs)
	}

	return runPluginCommand(args, pluginPath, getPluginLogs)
}

// getPluginLogs fetches the latest 500 log entries from Mattermost,
// and prints only the ones related to the plugin to stdout.
func getPluginLogs(ctx context.Context, client *model.Client4, pluginID string) error {
	Logger.Info("Getting plugin logs", "plugin_id", pluginID)

	err := checkJSONLogsSetting(ctx, client)
	if err != nil {
		return err
	}

	logs, err := fetchLogs(ctx, client, LogsRequest{
		Page:     0,
		PerPage:  defaultLogsPerPage,
		PluginID: pluginID,
		Since:    time.Unix(0, 0),
	})
	if err != nil {
		return fmt.Errorf("failed to fetch log entries: %w", err)
	}

	printLogEntries(logs)

	return nil
}

// watchPluginLogs fetches log entries from Mattermost and prints them continuously.
// It will return without an error when ctx is canceled.
func watchPluginLogs(ctx context.Context, client *model.Client4, pluginID string) error {
	Logger.Info("Watching plugin logs", "plugin_id", pluginID)

	err := checkJSONLogsSetting(ctx, client)
	if err != nil {
		return err
	}

	now := time.Now()
	var oldestEntry string

	// Use context.WithoutCancel to keep watching even if parent context times out
	watchCtx := context.WithoutCancel(ctx)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-watchCtx.Done():
			return nil
		case <-ticker.C:
			var page int
			for {
				logs, err := fetchLogs(watchCtx, client, LogsRequest{
					Page:     page,
					PerPage:  logsPerPage,
					PluginID: pluginID,
					Since:    now,
				})
				if err != nil {
					return fmt.Errorf("failed to fetch log entries: %w", err)
				}

				var allNew bool
				logs, oldestEntry, allNew = checkOldestEntry(logs, oldestEntry)

				printLogEntries(logs)

				if !allNew {
					// No more logs to fetch
					break
				}
				page++
			}
		}
	}
}

// checkOldestEntry checks if logs contains new log entries.
// It returns the filtered slice of log entries, the new oldest entry and whether or not all entries were new.
func checkOldestEntry(logs []string, oldest string) (filteredLogs []string, newOldest string, allNew bool) {
	if len(logs) == 0 {
		return nil, oldest, false
	}

	newOldestEntry := logs[len(logs)-1]

	i := slices.Index(logs, oldest)
	switch i {
	case -1:
		// Every log entry is new
		return logs, newOldestEntry, true
	case len(logs) - 1:
		// No new log entries
		return nil, oldest, false
	default:
		// Filter out oldest log entry
		return logs[i+1:], newOldestEntry, false
	}
}

// LogsRequest contains parameters for fetching logs.
type LogsRequest struct {
	Page     int
	PerPage  int
	PluginID string
	Since    time.Time
}

// fetchLogs fetches log entries from Mattermost
// and filters them based on pluginID and timestamp.
func fetchLogs(ctx context.Context, client *model.Client4, req LogsRequest) ([]string, error) {
	logs, _, err := client.GetLogs(ctx, req.Page, req.PerPage)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs from Mattermost: %w", err)
	}

	logs, err = filterLogEntries(logs, req.PluginID, req.Since)
	if err != nil {
		return nil, fmt.Errorf("failed to filter log entries: %w", err)
	}

	return logs, nil
}

// filterLogEntries filters a given slice of log entries by pluginID.
// It also filters out any entries which timestamps are older than since.
func filterLogEntries(logs []string, pluginID string, since time.Time) ([]string, error) {
	type logEntry struct {
		PluginID  string `json:"plugin_id"`
		Timestamp string `json:"timestamp"`
	}

	ret := make([]string, 0)

	for _, e := range logs {
		var le logEntry
		err := json.Unmarshal([]byte(e), &le)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal log entry into JSON: %w", err)
		}
		if le.PluginID != pluginID {
			continue
		}

		let, err := time.Parse(timeStampFormat, le.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("unknown timestamp format: %w", err)
		}
		if let.Before(since) {
			continue
		}

		// Log entries returned by the API have a newline as prefix.
		// Remove that to make printing consistent.
		e = strings.TrimPrefix(e, "\n")

		ret = append(ret, e)
	}

	return ret, nil
}

// printLogEntries prints a slice of log entries to stdout.
func printLogEntries(entries []string) {
	for _, e := range entries {
		fmt.Println(e)
	}
}

func checkJSONLogsSetting(ctx context.Context, client *model.Client4) error {
	cfg, _, err := client.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch config: %w", err)
	}
	if cfg.LogSettings.FileJson == nil || !*cfg.LogSettings.FileJson {
		return errors.New("JSON output for file logs are disabled. " +
			"Please enable LogSettings.FileJson via the configuration in Mattermost")
	}

	return nil
}
