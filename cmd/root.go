// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cmd

import (
	_ "net/http/pprof"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/joshuar/go-hass-agent/cmd/text"
	"github.com/joshuar/go-hass-agent/internal/agent"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

var (
	traceFlag    bool
	debugFlag    bool
	AppID        string
	profileFlag  bool
	headlessFlag bool
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "go-hass-agent",
	Short: "A Home Assistant, native app integration for desktop/laptop devices.",
	Long:  text.RootCmdLongText,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logging.SetLoggingLevel(traceFlag, debugFlag, profileFlag)
		logging.SetLogFile("go-hass-agent.log")
	},
	Run: func(cmd *cobra.Command, args []string) {
		agent := agent.New(&agent.Options{
			Headless: headlessFlag,
			ID:       AppID,
		})
		var err error

		var trk *sensor.SensorTracker
		if trk, err = sensor.NewSensorTracker(agent.AppID()); err != nil {
			log.Fatal().Err(err).Msg("Could not start sensor sensor.")
		}

		agent.Run(trk)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Msg("Could not start.")
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&traceFlag, "trace", false,
		"trace output (default is false)")
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false,
		"debug output (default is false)")
	rootCmd.PersistentFlags().BoolVar(&profileFlag, "profile", false,
		"enable profiling (default is false)")
	rootCmd.PersistentFlags().StringVar(&AppID, "appid", "com.github.joshuar.go-hass-agent",
		"specify a custom app ID (for debugging)")
	rootCmd.PersistentFlags().BoolVar(&headlessFlag, "terminal", defaultHeadless(),
		"run in terminal (without a GUI)")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(registerCmd)
}

func defaultHeadless() bool {
	_, v := os.LookupEnv("DISPLAY")
	return !v
}
