// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint
package main

import (
	"context"
	"fmt"
	"log/slog"

	pulseaudiox "github.com/joshuar/go-hass-agent/pkg/linux/pulseaudio"
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
			slog.Error("failed to parse reply: %w", err)
		}
		volPct := pulseaudiox.ParseVolume(repl)
		switch {
		case repl.Mute != client.Mute:
			fmt.Printf("mute changed to %v\n", repl.Mute)
			client.Mute = repl.Mute
		case volPct != client.Vol:
			fmt.Printf("volume changed to %.0f%%\n", volPct)
			client.Vol = volPct
		}
	}
}
