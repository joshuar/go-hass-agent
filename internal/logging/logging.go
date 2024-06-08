// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package logging

import (
	"fmt"
	_ "net/http/pprof" // #nosec G108
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

//nolint:exhaustruct,reassign
func init() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

// SetLoggingLevel sets an appropriate log level and enables profiling if requested.
func SetLoggingLevel(level string) {
	switch level {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
		log.Debug().Msg("Trace logging enabled.")
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("Debug logging enabled.")
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// SetLogFile will attempt to create and then write logging to a file. If it
// cannot do this, logging will only be available on stdout.
//
//nolint:exhaustruct,mnd
func SetLogFile(filename string) error {
	logFile := filepath.Join(xdg.StateHome, filename)

	if err := checkPath(xdg.StateHome); err != nil {
		return fmt.Errorf("unable to create directory for log file: %w", err)
	}

	logWriter, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("unable to open log file: %w", err)
	}

	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
	multiWriter := zerolog.MultiLevelWriter(consoleWriter, logWriter)
	log.Logger = log.Output(multiWriter)

	return nil
}

// Reset will remove the log file.
func Reset() error {
	logFile := filepath.Join(xdg.StateHome, "go-hass-agent.log")

	err := os.Remove(logFile)
	if err != nil {
		return fmt.Errorf("could not remove log file: %w", err)
	}

	return nil
}

func checkPath(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return fmt.Errorf("unable to create path: %w", err)
		}
	}

	return nil
}
