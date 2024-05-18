// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package power

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	suspend powerSignal = iota
	shutdown

	sleepSignal    = managerInterface + ".PrepareForSleep"
	shutdownSignal = managerInterface + ".PrepareForShutdown"
)

type powerSignal int

type powerStateSensor struct {
	linux.Sensor
	signal powerSignal
}

func (s *powerStateSensor) State() any {
	b, ok := s.Value.(bool)
	if !ok {
		return sensor.StateUnknown
	}
	if b {
		switch s.signal {
		case suspend:
			return "Suspended"
		case shutdown:
			return "Powered Off"
		}
	}
	return "Powered On"
}

func (s *powerStateSensor) Icon() string {
	str, ok := s.State().(string)
	if !ok {
		str = ""
	}
	switch str {
	case "Suspended":
		return "mdi:power-sleep"
	case "Powered Off":
		return "mdi:power-off"
	default:
		return "mdi:power-on"
	}
}

func newPowerState(s powerSignal, v any) *powerStateSensor {
	return &powerStateSensor{
		signal: s,
		Sensor: linux.Sensor{
			SensorTypeValue: linux.SensorPowerState,
			Value:           v,
			SensorSrc:       linux.DataSrcDbus,
			IsDiagnostic:    true,
		},
	}
}

func StateUpdater(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details)

	events, err := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Watch(ctx, dbusx.Watch{
			Names:     []string{sleepSignal, shutdownSignal},
			Interface: managerInterface,
			Path:      loginBasePath,
		})
	if err != nil {
		log.Debug().Err(err).
			Msg("Failed to create power state D-Bus watch.")
		close(sensorCh)
		return sensorCh
	}

	go func() {
		defer close(sensorCh)
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("Stopped power state sensor.")
				return
			case event := <-events:
				switch event.Signal {
				case sleepSignal:
					go func() {
						sensorCh <- newPowerState(suspend, event.Content[0])
					}()
				case shutdownSignal:
					go func() {
						sensorCh <- newPowerState(shutdown, event.Content[0])
					}()
				}
			}
		}
	}()

	// Send an initial sensor update.
	go func() {
		sensorCh <- newPowerState(shutdown, false)
	}()

	return sensorCh
}
