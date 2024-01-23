// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cmd

import (
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/joshuar/go-hass-agent/internal/agent"
	"github.com/joshuar/go-hass-agent/internal/agent/config"
	"github.com/joshuar/go-hass-agent/internal/tracker"
)

var (
	serverFlag, tokenFlag string
	forcedFlag            bool
)

// registerCmd represents the register command.
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register this device with Home Assistant",
	Long:  `Register will attempt to register this device with Home Assistant. A URL for a Home Assistant instance and long-lived access token can be provided if known beforehand.`,
	Run: func(cmd *cobra.Command, args []string) {
		agent := agent.New(&agent.Options{
			Headless:      headlessFlag,
			ForceRegister: forcedFlag,
			Server:        serverFlag,
			Token:         tokenFlag,
			ID:            appID,
		})
		var err error
		var cfg config.Config
		configPath := filepath.Join(xdg.ConfigHome, agent.AppID())
		if cfg, err = config.Load(configPath); err != nil {
			log.Fatal().Err(err).Msg("Could not load config.")
		}
		var trk *tracker.SensorTracker
		if trk, err = tracker.NewSensorTracker(agent.AppID()); err != nil {
			log.Fatal().Err(err).Msg("Could not start sensor tracker.")
		}
		agent.Register(cfg, trk)
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
