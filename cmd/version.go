/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	Name    = "go-hass-agent"
	Version = "0.0.1"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: fmt.Sprintf("Print the version number of %s", Name),
	Long:  fmt.Sprintf("All software has versions. This is the version of %s", Name),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s: %s", Name, Version)
	},
}
