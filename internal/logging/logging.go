// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package logging

import (
	"fmt"
	"net/http"
	"os"

	_ "net/http/pprof"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func init() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func setProfiling() {
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

// SetLogging sets an appropriate log level and enables profiling if requested.
func SetLogging(trace, debug, profile bool) {
	switch {
	case trace:
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
		log.Debug().Msg("Trace logging enabled.")
	case debug:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("Debug logging enabled.")
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	if profile {
		setProfiling()
	}
}
