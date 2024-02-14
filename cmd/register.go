// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/joshuar/go-hass-agent/cmd/text"
	"github.com/joshuar/go-hass-agent/internal/agent"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

var (
	serverFlag, tokenFlag string
	forcedFlag            bool
)

// registerCmd represents the register command.
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register this device with Home Assistant",
	Long:  text.RegisterCmdLongText,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logging.SetLoggingLevel(traceFlag, debugFlag, profileFlag)
		logging.SetLogFile("go-hass-agent.log")
	},
	Run: func(cmd *cobra.Command, args []string) {
		agent := agent.New(&agent.Options{
			Headless:      headlessFlag,
			ForceRegister: forcedFlag,
			Server:        serverFlag,
			Token:         tokenFlag,
			ID:            AppID,
		})
		var err error

		var trk *sensor.SensorTracker
		if trk, err = sensor.NewSensorTracker(agent.AppID()); err != nil {
			log.Fatal().Err(err).Msg("Could not start sensor sensor.")
		}

		agent.Register(trk)
	},
}

func init() {
	registerCmd.PersistentFlags().StringVar(&serverFlag,
		"server", "http://localhost:8123",
		"URL to Home Assistant instance (e.g. https://somehost:someport)")
	registerCmd.PersistentFlags().StringVar(&tokenFlag,
		"token", "",
		"Long-lived token (e.g. 123456)")
	registerCmd.PersistentFlags().BoolVar(&forcedFlag,
		"force", false,
		"Ignore any previous registration and re-register the agent.")
}
