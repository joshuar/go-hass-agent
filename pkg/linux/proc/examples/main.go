// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"fmt"
	"log/slog"
	"os"

	diskstats "github.com/joshuar/go-hass-agent/pkg/linux/proc"
)

func main() {
	allStats, err := diskstats.ReadDiskStatsFromSysFS()
	if err != nil {
		slog.Error("Failed to read /proc/stats", "error", err)
		os.Exit(-1)
	}
	for dev, stats := range allStats {
		fmt.Fprintf(os.Stdout, "Device: %s\n", dev)
		for k, v := range stats {
			fmt.Fprintf(os.Stdout, "\t%s: %d\n", k.String(), v)
		}
	}
}
