/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/joshuar/go-hass-agent/internal/device"
	hass "github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(infoCmd)
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Print details of this device",
	Long:  "This will show the information that was used to register this device with Home Assistant",
	Run: func(cmd *cobra.Command, args []string) {
		device := device.NewDevice()
		hass.GetDeviceInfo(device)
	},
}
