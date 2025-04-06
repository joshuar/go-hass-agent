// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Package logging contains methods and object for handling logging in the agent.
package logging

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	slogmulti "github.com/samber/slog-multi"
)

const (
	// LevelTrace is a custom log level representing trace-level logs.
	LevelTrace = slog.Level(-8)
	// LevelFatal is a custom log level representing fatal-level logs.
	LevelFatal = slog.Level(12)

	logFileName = "agent.log"
)

// ErrLogOption represents an error or problem with a logging option.
var ErrLogOption = errors.New("logging option error")

// LevelNames is a map of custom log level names to their string values.
var LevelNames = map[slog.Leveler]string{
	LevelTrace: "TRACE",
	LevelFatal: "FATAL",
}

// Options contains the top-level logging options.
type Options struct {
	LogLevel  string `name:"log-level" enum:"info,debug,trace" default:"info" help:"Set logging level."`
	NoLogFile bool   `name:"no-log-file" help:"Don't write to a log file." default:"false"`
	Path      string `kong:"-"`
}

// New creates a new logger with the given options.
func New(options Options) *slog.Logger {
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
		logFile = filepath.Join(options.Path, logFileName)
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
		err = os.MkdirAll(logDir, os.ModeAppend)
		if err != nil {
			return nil, fmt.Errorf("unable to create log file directory %s: %w", logDir, err)
		}
	}

	// Open the log file.
	logFileHandle, err := os.Create(logFile) // #nosec G304
	if err != nil {
		return nil, fmt.Errorf("unable to open log file: %w", err)
	}

	return logFileHandle, nil
}

// Reset will remove the log file.
func Reset(path string) error {
	logFile := filepath.Join(path, logFileName)

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
