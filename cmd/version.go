// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cmd

import (
	"github.com/joshuar/go-hass-agent/internal/agent"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setLogfileLogging()
		setLoggingLevel()
		setProfiling()
	},
	Run: func(cmd *cobra.Command, args []string) {
		agent.ShowVersion(agent.AgentOptions{ID: appID})
	},
}
