// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog/log"

	diskstats "github.com/joshuar/go-hass-agent/pkg/linux/proc"
)

func main() {
	stats, err := diskstats.ReadDiskstats()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not read.")
	}
	spew.Dump(stats)
}
