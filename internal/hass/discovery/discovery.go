// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Package discovery implements local network discovery of Home Assistant servers.
package discovery

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/grandcat/zeroconf"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
)

const (
	haDiscoveryTimeout = 5 * time.Second
)

// FindServers is a helper function to generate a list of Home Assistant servers
// via local network auto-discovery.
func FindServers(ctx context.Context) ([]string, error) {
	serverList := []string{preferences.DefaultServer}

	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return serverList, fmt.Errorf("failed to initialize resolver: %w", err)
	}

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
				serverList = append(serverList, server)
			} else {
				slogctx.FromCtx(ctx).Log(ctx, logging.LevelTrace,
					"Found a server malformed server, will not use.", slog.String("server", entry.HostName))
			}
		}
	}(entries)

	slogctx.FromCtx(ctx).Info("Looking for Home Assistant servers on the local network...")

	searchCtx, searchCancel := context.WithTimeout(ctx, haDiscoveryTimeout)
	defer searchCancel()

	err = resolver.Browse(searchCtx, "_home-assistant._tcp", "local.", entries)
	if err != nil {
		return serverList, fmt.Errorf("could not start search for Home Assistant servers: %w", err)
	}

	<-searchCtx.Done()

	return serverList, nil
}
