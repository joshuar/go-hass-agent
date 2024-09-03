// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package logging

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"

	slogmulti "github.com/samber/slog-multi"
)

const (
	LevelTrace = slog.Level(-8)
	LevelFatal = slog.Level(12)
)

var LevelNames = map[slog.Leveler]string{
	LevelTrace: "TRACE",
	LevelFatal: "FATAL",
}

//revive:disable:flag-parameter
func New(level string, logFile string) *slog.Logger {
	var (
		logLevel slog.Level
		handler  slog.Handler
	)

	// Set the log level.
	switch level {
	case "trace":
		logLevel = LevelTrace
	case "debug":
		logLevel = slog.LevelDebug
	default:
		logLevel = slog.LevelInfo
	}

	// Set the slog handler
	// Unless no log file was requested, set up file logging.
	if logFile != "" {
		logFH, err := openLogFile(logFile)
		if err != nil {
			slog.Warn("unable to open log file",
				slog.String("file", logFile),
				slog.Any("error", err))
		} else {
			handler = slogmulti.Fanout(
				tint.NewHandler(os.Stdout, generateOptions(logLevel, os.Stdout.Fd())),
				tint.NewHandler(logFH, generateOptions(logLevel, logFH.Fd())),
			)
		}
	} else {
		handler = slogmulti.Fanout(
			tint.NewHandler(os.Stdout, generateOptions(logLevel, os.Stdout.Fd())),
		)
	}

	logger := slog.New(handler)

	slog.SetDefault(logger)

	return logger
}

func generateOptions(level slog.Level, fd uintptr) *tint.Options {
	opts := &tint.Options{
		Level:   level,
		NoColor: !isatty.IsTerminal(fd),
	}
	if level == LevelTrace {
		opts.AddSource = true
	}

	return opts
}

//nolint:unused
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
func Reset(file string) error {
	// If the log file doesn't exist, just exit.
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		return nil
	}
	// Else, remove the file.
	err = os.Remove(file)
	if err != nil {
		return fmt.Errorf("could not remove log file: %w", err)
	}

	return nil
}
