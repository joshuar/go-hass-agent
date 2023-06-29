// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cmd

import (
	"github.com/joshuar/go-hass-agent/internal/agent"
	"github.com/spf13/cobra"
)

var (
	serverFlag, tokenFlag string
	forcedFlag            bool
	args                  []string
)

// registerCmd represents the register command
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register this device with Home Assistant",
	Long:  `Register will attempt to register this device with Home Assistant. A URL for a Home Assistant instance and long-lived access token can be provided if known beforehand.`,
	Run: func(cmd *cobra.Command, args []string) {
		agent.Register(agent.AgentOptions{
			Headless: headlessFlag,
			Register: forcedFlag,
			ID:       appID,
		}, serverFlag, tokenFlag)
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
	rootCmd.AddCommand(registerCmd)
}
