// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package logging

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	slogmulti "github.com/samber/slog-multi"

	"github.com/joshuar/go-hass-agent/internal/preferences"

	"github.com/adrg/xdg"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

const (
	LevelTrace = slog.Level(-8)
	LevelFatal = slog.Level(12)

	logFileName = "agent.log"
)

var LevelNames = map[slog.Leveler]string{
	LevelTrace: "TRACE",
	LevelFatal: "FATAL",
}

type Options struct {
	LogLevel  string `name:"log-level" enum:"info,debug,trace" default:"info" help:"Set logging level."`
	NoLogFile bool   `name:"no-log-file" help:"Don't write to a log file." default:"false"`
}

//revive:disable:flag-parameter
func New(appID string, options Options) *slog.Logger {
	var (
		logLevel slog.Level
		logFile  string
		handlers []slog.Handler
	)

	// Set the log level.
	switch options.LogLevel {
	case "trace":
		logLevel = LevelTrace
	case "debug":
		logLevel = slog.LevelDebug
	default:
		logLevel = slog.LevelInfo
	}

	// Set a log file if specified.
	if options.NoLogFile {
		logFile = ""
	} else {
		logFile = filepath.Join(xdg.ConfigHome, appID, logFileName)
	}

	handlers = append(handlers, tint.NewHandler(os.Stderr,
		generateConsoleOptions(logLevel, os.Stderr.Fd())))

	// Unless no log file was requested, set up file logging.
	if logFile != "" {
		logFH, err := openLogFile(logFile)
		if err != nil {
			slog.Warn("unable to open log file",
				slog.String("file", logFile),
				slog.Any("error", err))
		} else {
			handlers = append(handlers, slog.NewTextHandler(logFH, generateFileOpts(logLevel)))
		}
	}

	logger := slog.New(slogmulti.Fanout(handlers...))
	slog.SetDefault(logger)

	return logger
}

func generateConsoleOptions(level slog.Level, fd uintptr) *tint.Options {
	opts := &tint.Options{
		Level:       level,
		NoColor:     !isatty.IsTerminal(fd),
		ReplaceAttr: tintLevelReplacer,
		TimeFormat:  time.Kitchen,
	}
	if level == LevelTrace {
		opts.AddSource = true
	}

	return opts
}

func generateFileOpts(level slog.Level) *slog.HandlerOptions {
	opts := &slog.HandlerOptions{
		Level:       level,
		ReplaceAttr: fileLevelReplacer,
	}
	if level == LevelTrace {
		opts.AddSource = true
	}

	return opts
}

func tintLevelReplacer(_ []string, attr slog.Attr) slog.Attr {
	// Set default level.
	if attr.Key == slog.LevelKey {
		level, ok := attr.Value.Any().(slog.Level)
		if !ok {
			level = slog.LevelInfo
		}

		// Errors in red.
		if err, ok := attr.Value.Any().(error); ok {
			aErr := tint.Err(err)
			attr.Key = aErr.Key
		}

		// Format custom log level.
		levelLabel, exists := LevelNames[level]
		if exists {
			attr.Value = slog.StringValue(levelLabel)
		}
	}

	return attr
}

func fileLevelReplacer(_ []string, attr slog.Attr) slog.Attr {
	// Set default level.
	if attr.Key == slog.LevelKey {
		level, ok := attr.Value.Any().(slog.Level)
		if !ok {
			level = slog.LevelInfo
		}

		// Format custom log level.
		levelLabel, exists := LevelNames[level]
		if exists {
			attr.Value = slog.StringValue(levelLabel)
		}
	}

	return attr
}

// openLogFile will attempt to open the specified log file. It will also attempt
// to create the directory containing the log file if it does not exist.
func openLogFile(logFile string) (*os.File, error) {
	logDir := filepath.Dir(logFile)
	// Create the log directory if it does not exist.
	_, err := os.Stat(logDir)

	if err == nil || errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(logDir, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("unable to create log file directory %s: %w", logDir, err)
		}
	}

	// Open the log file.
	logFileHandle, err := os.Create(logFile)
	if err != nil {
		return nil, fmt.Errorf("unable to open log file: %w", err)
	}

	return logFileHandle, nil
}

// Reset will remove the log file.
func Reset(ctx context.Context) error {
	appID := preferences.AppIDFromContext(ctx)
	logFile := filepath.Join(xdg.ConfigHome, appID, logFileName)

	// If the log file doesn't exist, just exit.
	_, err := os.Stat(logFile)
	if os.IsNotExist(err) {
		return nil
	}
	// Else, remove the file.
	err = os.Remove(logFile)
	if err != nil {
		return fmt.Errorf("could not remove log file: %w", err)
	}

	return nil
}
