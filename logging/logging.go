// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package logging

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	slogmulti "github.com/samber/slog-multi"
	slogctx "github.com/veqryn/slog-context"
	slogjson "github.com/veqryn/slog-json"

	"github.com/joshuar/go-hass-agent/config"
)

// ErrLogOption indicates an invalid logging option.
var ErrLogOption = errors.New("invalid logging option")

// LevelNames contains a list of custom log level names.
var LevelNames = map[slog.Leveler]string{
	LevelTrace: "TRACE",
	LevelFatal: "FATAL",
}

const (
	// LevelTrace is a custom TRACE log level.
	LevelTrace = slog.Level(-8)
	// LevelFatal is a custom FATAL log level.
	LevelFatal = slog.Level(12)
)

// Options are options for controlling logging.
type Options struct {
	LogLevel  string `env:"GOHASSAGENT_LOGLEVEL"  name:"log-level"   enum:"info,debug,trace" default:"info"  help:"Set logging level."`
	NoLogFile bool   `env:"GOHASSAGENT_NOLOGFILE" name:"no-log-file"                         default:"false" help:"Don't write to a log file."`
}

// New creates a new logger with the given options.
func New(options Options) {
	var (
		logFile string
		level   slog.Level
	)
	// Set the log level.
	switch options.LogLevel {
	case "trace":
		level = LevelTrace
	case "debug":
		level = slog.LevelDebug
	default:
		level = slog.LevelInfo
	}
	// Set a log file if specified.
	if options.NoLogFile {
		logFile = ""
	} else {
		logFile = filepath.Join(config.GetPath(), "go-hass-agent.log")
	}

	// Set up handlers.
	var handlers []slog.Handler
	handlers = append(handlers,
		tint.NewHandler(os.Stderr, generateConsoleOptions(level, os.Stderr.Fd())),
	)
	// Unless no log file was requested, set up file logging.
	if logFile != "" {
		logFH, err := openLogFile(logFile)
		if err != nil {
			slog.Warn("unable to open log file",
				slog.String("file", logFile),
				slog.Any("error", err))
		} else {
			handlers = append(handlers,
				slogjson.NewHandler(logFH, generateFileOpts(level)),
			)
		}
	}

	slog.SetDefault(slog.New(slogctx.NewHandler(slogmulti.Fanout(handlers...), nil)))
}

func generateConsoleOptions(level slog.Level, fd uintptr) *tint.Options {
	opts := &tint.Options{
		Level:       level,
		NoColor:     !isatty.IsTerminal(fd),
		ReplaceAttr: consolelevelReplacer,
		TimeFormat:  time.Kitchen,
	}
	if level == LevelTrace {
		opts.AddSource = true
	}

	return opts
}

func generateFileOpts(level slog.Level) *slogjson.HandlerOptions {
	opts := &slogjson.HandlerOptions{
		AddSource:   false,
		Level:       level,
		ReplaceAttr: fileLevelReplacer,
	}
	if level == LevelTrace {
		opts.AddSource = true
	}

	return opts
}

func consolelevelReplacer(_ []string, attr slog.Attr) slog.Attr {
	if attr.Key == slog.LevelKey {
		level, ok := attr.Value.Any().(slog.Level)
		if !ok {
			level = slog.LevelInfo
		}
		switch level {
		case slog.LevelError:
			attr.Value = slog.StringValue(color.HiRedString("ERROR"))
		case slog.LevelWarn:
			attr.Value = slog.StringValue(color.HiYellowString("WARN"))
		case slog.LevelInfo:
			attr.Value = slog.StringValue(color.HiGreenString("INFO"))
		case slog.LevelDebug:
			attr.Value = slog.StringValue(color.HiMagentaString("DEBUG"))
		case LevelTrace:
			attr.Value = slog.StringValue(color.HiWhiteString("TRACE"))
		default:
			attr.Value = slog.StringValue("UNKNOWN")
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
		err = os.MkdirAll(logDir, 0o750)
		if err != nil {
			return nil, fmt.Errorf("unable to create log file directory %s: %w", logDir, err)
		}
	}

	// Open the log file.
	logFileHandle, err := os.Create(logFile) // #nosec:G304
	if err != nil {
		return nil, fmt.Errorf("unable to open log file: %w", err)
	}

	return logFileHandle, nil
}
