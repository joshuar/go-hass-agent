/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"net/http"
	"os"

	"github.com/joshuar/go-hass-agent/internal/agent"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cfgFile     string
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
		if debugFlag {
			log.SetLevel(log.DebugLevel)
			log.Debug("Debug logging enabled")
		}
		if profileFlag {
			go func() {
				log.Info(http.ListenAndServe("localhost:6060", nil))
			}()
			log.Info("Profiling is enabled and available at localhost:6060")
		}
		agent := agent.NewAgent()
		agent.GetRegistrationInfo()

	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.go-hass-agent.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	// cobra.OnInitialize(config.InitialiseConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "$HOME/.config/go-hass-app.yaml", "config file (default is $HOME/.config/go-hass-app.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&debugFlag, "debug", "d", false, "debug output (default is false)")
	rootCmd.PersistentFlags().BoolVarP(&profileFlag, "profile", "p", false, "enable profiling (default is false)")
}
