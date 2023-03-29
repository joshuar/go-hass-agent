/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/joshuar/go-hass-agent/internal/agent"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	debugFlag   bool
	profileFlag bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-hass-agent",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
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
		agent := agent.NewAgent()
		agent.App.Run()
		agent.Exit()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debugFlag, "debug", "d", false, "debug output (default is false)")
	rootCmd.PersistentFlags().BoolVarP(&profileFlag, "profile", "p", false, "enable profiling (default is false)")
}
