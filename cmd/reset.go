/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/joshuar/go-hass-agent/cmd/text"
	"github.com/joshuar/go-hass-agent/internal/agent"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/registry"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

// resetCmd represents the reset command
var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset (remove) the configuration and log files.",
	Long:  text.ResetCmdLongText,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logging.SetLoggingLevel(traceFlag, debugFlag, profileFlag)
		if !noLogFileFlag {
			logging.SetLogFile("go-hass-agent.log")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		agent := agent.New(&agent.Options{
			Headless: headlessFlag,
			ID:       AppID,
		})
		registry.SetPath(filepath.Join(xdg.ConfigHome, agent.AppID(), "sensorRegistry"))
		preferences.SetPath(filepath.Join(xdg.ConfigHome, agent.AppID()))
		// Reset agent.
		if err := agent.Reset(); err != nil {
			log.Warn().Err(err).Msg("Could not reset agent.")
		}
		// Reset registry.
		if err := registry.Reset(); err != nil {
			log.Warn().Err(err).Msg("Could not reset registry.")
		}
		// Reset preferences.
		if err := preferences.Reset(); err != nil {
			log.Warn().Err(err).Msg("Could not reset preferences.")
		}
		// Reset the log.
		if err := logging.Reset(); err != nil {
			log.Warn().Err(err).Msg("Could not remove log file.")
		}
		log.Info().Msg("Reset complete (refer to any warnings, if any, above.)")
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)
}
