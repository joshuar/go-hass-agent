// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func setLogging() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func setDebugging() {
	if debugFlag {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("Debug logging enabled.")
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

func setProfiling() {
	if profileFlag {
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
}
