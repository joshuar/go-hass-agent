// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cmd

import (
	"github.com/joshuar/go-hass-agent/internal/agent"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Print details of this device",
	Long:  "This will show the information that was used to register this device with Home Assistant",
	Run: func(cmd *cobra.Command, args []string) {
		agent.ShowInfo(agent.Options{ID: appID})
	},
}
