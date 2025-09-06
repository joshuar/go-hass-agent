// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"context"
	"log/slog"

	"github.com/joshuar/go-hass-agent/pkg/linux/pulseaudiox"
)

func main() {
	client, err := pulseaudiox.NewPulseClient(context.Background())
	if err != nil {
		panic(err)
	}

	err = client.SetVolume(20)
	if err != nil {
		panic(err)
	}

	for {
		<-client.EventCh

		repl, err := client.GetState()
		if err != nil {
			slog.Error("failed to parse reply: %w", slog.Any("error", err))
		}

		volPct := pulseaudiox.ParseVolume(repl)

		switch {
		case repl.Mute != client.Mute:
			slog.Info("mute changed", slog.Bool("state", repl.Mute))
			client.Mute = repl.Mute
		case volPct != client.Vol:
			slog.Info("volume changed.", slog.Float64("state", volPct))
			client.Vol = volPct
		}
	}
}
