// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cmd

import (
	"context"
	"net/url"
	"os"

	"fyne.io/fyne/v2/data/binding"
	"github.com/joshuar/go-hass-agent/internal/agent"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	providedURL, providedToken string
)

// registerCmd represents the register command
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register this device with Home Assistant",
	Long:  `Register will attempt to register this device with Home Assistant. A URL for a Home Assistant instance and long-lived access token is required to be provided.`,
	Run: func(cmd *cobra.Command, args []string) {
		agentCtx, cancelFunc, agent := agent.NewAgent("")
		log.Info().Msg("Starting registration process.")
		agent.SetupLogging()
		agent.CheckConfig(agentCtx, registrationFetcher)
		cancelFunc()
		log.Info().Msg("Device registered with Home Assistant.")
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
	registerCmd.PersistentFlags().StringVar(&providedURL,
		"url", "http://localhost:8123",
		"URL to Home Assistant instance (e.g. https://somehost:someport)")
	registerCmd.PersistentFlags().StringVar(&providedToken,
		"token", "",
		"Long-lived token (e.g. 123456)")
	registerCmd.MarkPersistentFlagRequired("token")
}

func registrationFetcher(ctx context.Context) *hass.RegistrationHost {
	u, err := url.Parse(providedURL)
	if err != nil {
		log.Error().Err(err).
			Msg("Cannot parse provided URL.")
		os.Exit(-1)
	}
	registrationInfo := agent.NewRegistration()
	registrationInfo.Token = binding.BindString(&providedToken)
	registrationInfo.Server = binding.BindString(&u.Host)
	registrationInfo.UseTLS = binding.NewBool()
	if u.Scheme == "https" {
		registrationInfo.UseTLS.Set(true)
	} else {
		registrationInfo.UseTLS.Set(false)
	}
	return registrationInfo
}
