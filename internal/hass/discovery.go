// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"
	"strings"
	"time"

	"fyne.io/fyne/v2/data/binding"
	"github.com/grandcat/zeroconf"
	"github.com/rs/zerolog/log"
)

// FindServers is a helper function to generate a list of Home Assistant servers
// via local network auto-discovery.
func FindServers(ctx context.Context) binding.StringList {
	serverList := binding.NewStringList()

	// add http://localhost:8123 to the list of servers as a fall-back/default
	// option
	if err := serverList.Append("http://localhost:8123"); err != nil {
		log.Debug().Err(err).
			Msg("Unable to set a default server.")
	}

	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to initialize resolver.")
	} else {
		entries := make(chan *zeroconf.ServiceEntry)
		go func(results <-chan *zeroconf.ServiceEntry) {
			for entry := range results {
				var server string
				for _, t := range entry.Text {
					if value, found := strings.CutPrefix(t, "base_url="); found {
						server = value
					}
				}
				if server != "" {
					if err = serverList.Append(server); err != nil {
						log.Warn().Err(err).
							Msgf("Unable to add found server %s to server list.", server)
					}
				} else {
					log.Debug().Msgf("Entry %s did not have a base_url value. Not using it.", entry.HostName)
				}
			}
		}(entries)

		log.Info().Msg("Looking for Home Assistant instances on the network...")
		searchCtx, searchCancel := context.WithTimeout(ctx, time.Second*5)
		defer searchCancel()
		err = resolver.Browse(searchCtx, "_home-assistant._tcp", "local.", entries)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to browse")
		}

		<-searchCtx.Done()
	}
	return serverList
}
