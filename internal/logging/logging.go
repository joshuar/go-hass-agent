// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package logging

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func init() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func SetProfiling() {
	go func() {
		for i := 6060; i < 6070; i++ {
			log.Debug().
				Msgf("Starting profiler web interface on localhost:" + fmt.Sprint(i))
			err := http.ListenAndServe("localhost:"+fmt.Sprint(i), nil)
			if err != nil {
				log.Debug().Err(err).
					Msg("Trouble starting profiler, trying again.")
			}
		}
	}()
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
func SetLogFile(filename string) {
	logFile := filepath.Join(xdg.StateHome, filename)
	if err := checkPath(xdg.StateHome); err != nil {
		log.Warn().Err(err).Msg("Unable to create directory for log file storage. Will not log to disk.")
		return
	}
	logWriter, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Warn().Err(err).
			Msg("Unable to open log file for writing. Will not log to disk.")
	} else {
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
		multiWriter := zerolog.MultiLevelWriter(consoleWriter, logWriter)
		log.Logger = log.Output(multiWriter)
	}
}

// Reset will remove the log file.
func Reset() error {
	logFile := filepath.Join(xdg.StateHome, "go-hass-agent.log")
	return os.Remove(logFile)
}

func checkPath(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.MkdirAll(path, os.ModePerm)
	}
	return nil
}
