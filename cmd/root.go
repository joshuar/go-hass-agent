// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cmd

import (
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/joshuar/go-hass-agent/internal/agent"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
)

var (
	debugFlag    bool
	debugID      string
	profileFlag  bool
	headlessFlag bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-hass-agent",
	Short: "A Home Assistant, native app integration for desktop/laptop devices.",
	Long: `go-hass-agent reports various sensors from a desktop/laptop to a Home Assistant instance. Sensors include the usual system metrics like load average and memory usage as well as things like current active app where possible.
	
	It can also receive notifications from Home Assistant.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		if debugFlag {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
			log.Debug().Msg("Debug logging enabled.")
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}
		if profileFlag {
			go func() {
				log.Info().Err(http.ListenAndServe("localhost:6060", nil))
			}()
			log.Info().Msg("Profiling is enabled and available at localhost:6060.")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		var runCmd func(string)
		if headlessFlag {
			runCmd = agent.RunHeadless
		} else {
			runCmd = agent.Run
		}
		if debugID != "" {
			runCmd(debugID)
		} else {
			runCmd("")
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not start.")
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debugFlag, "debug", "d", false,
		"debug output (default is false)")
	rootCmd.PersistentFlags().BoolVarP(&profileFlag, "profile", "p", false,
		"enable profiling (default is false)")
	rootCmd.PersistentFlags().StringVar(&debugID, "debugID", "",
		"specify a custom app ID (for debugging)")
	rootCmd.Flags().BoolVarP(&headlessFlag, "terminal", "t", false,
		"run in terminal (without a GUI)")
}
