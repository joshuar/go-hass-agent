// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cmd

import (
	_ "net/http/pprof"

	"github.com/joshuar/go-hass-agent/internal/agent"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	traceFlag    bool
	debugFlag    bool
	debugID      string
	profileFlag  bool
	headlessFlag bool
)

var appID = "com.github.joshuar.go-hass-agent"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-hass-agent",
	Short: "A Home Assistant, native app integration for desktop/laptop devices.",
	Long: `go-hass-agent reports various sensors from a desktop/laptop to a Home Assistant instance. Sensors include the usual system metrics like load average and memory usage as well as things like current active app where possible.
	
	It can also receive notifications from Home Assistant.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setLogfileLogging()
		setLoggingLevel()
		setProfiling()
	},
	Run: func(cmd *cobra.Command, args []string) {
		agent.Run(agent.AgentOptions{
			Headless: headlessFlag,
			ID:       appID,
		})
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
	rootCmd.PersistentFlags().StringVar(&debugID, "debugID", "",
		"specify a custom app ID (for debugging)")
	rootCmd.PersistentFlags().BoolVar(&headlessFlag, "terminal", false,
		"run in terminal (without a GUI)")
}
