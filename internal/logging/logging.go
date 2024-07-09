// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package logging

import (
	"fmt"
	"log/slog"
	_ "net/http/pprof" // #nosec G108
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/lmittmann/tint"
	slogmulti "github.com/samber/slog-multi"
)

const (
	LevelTrace = slog.Level(-8)
	LevelFatal = slog.Level(12)
)

//nolint:misspell
var LevelNames = map[slog.Leveler]string{
	LevelTrace: "TRACE",
	LevelFatal: "FATAL",
}

var DefaultLogFile = filepath.Join(xdg.ConfigHome, "go-hass-agent.log")

//nolint:exhaustruct
//revive:disable:flag-parameter
func New(level string, noLogFile bool) *slog.Logger {
	// Create an slog consoleOpts object.
	consoleOpts := &tint.Options{
		ReplaceAttr: levelReplacer,
	}
	fileOpts := &tint.Options{
		ReplaceAttr: levelReplacer,
		NoColor:     true,
	}

	// Set the log level.
	switch level {
	case "trace":
		consoleOpts.Level = LevelTrace
		consoleOpts.AddSource = true
	case "debug":
		consoleOpts.Level = slog.LevelDebug
	default:
		consoleOpts.Level = slog.LevelInfo
	}

	var logFile string
	if !noLogFile {
		logFile = DefaultLogFile
	}
	// Set the slog handler
	logHandler := slogmulti.Fanout(
		tint.NewHandler(os.Stdout, consoleOpts),
	)
	// Unless no log file was requested, set up file logging.
	if logFile != "" {
		logFile, err := openLogFile(logFile)
		if err != nil {
			slog.Warn("unable to open log file", "file", logFile, "error", err)
		} else {
			logHandler = slogmulti.Fanout(
				tint.NewHandler(os.Stdout, consoleOpts),
				tint.NewHandler(logFile, fileOpts),
			)
		}
	}

	logger := slog.New(logHandler)

	slog.SetDefault(logger)

	return logger
}

func levelReplacer(_ []string, attr slog.Attr) slog.Attr {
	if attr.Key == slog.LevelKey {
		level, ok := attr.Value.Any().(slog.Level)
		if !ok {
			level = slog.LevelInfo
		}

		levelLabel, exists := LevelNames[level]
		if !exists {
			levelLabel = level.String()
		}

		attr.Value = slog.StringValue(levelLabel)
	}

	return attr
}

//nolint:mnd
func openLogFile(logFile string) (*os.File, error) {
	logFileHandle, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("unable to open log file: %w", err)
	}

	return logFileHandle, nil
}

// Reset will remove the log file.
func Reset() error {
	err := os.Remove(DefaultLogFile)
	if err != nil {
		return fmt.Errorf("could not remove log file: %w", err)
	}

	return nil
}
