// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

// revive:disable:unused-receiver
package power

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	idleIcon    = "mdi:sleep"
	notIdleIcon = "mdi:sleep-off"

	idleProp     = managerInterface + "." + sessionIdleProp
	idleTimeProp = managerInterface + "." + sessionIdleTimeProp

	idleSF = 1000
)

type idleSensor struct {
	linux.Sensor
	idleTime int64
}

func (s *idleSensor) Icon() string {
	state, ok := s.State().(bool)
	if !ok {
		return notIdleIcon
	}

	switch state {
	case true:
		return idleIcon
	default:
		return notIdleIcon
	}
}

func (s *idleSensor) Attributes() any {
	return struct {
		DataSource string  `json:"data_source"`
		Seconds    float64 `json:"duration"`
	}{
		DataSource: linux.DataSrcDbus,
		Seconds:    idleTime(s.idleTime),
	}
}

//nolint:exhaustruct
func newIdleSensor(ctx context.Context) (*idleSensor, error) {
	newSensor := &idleSensor{
		Sensor: linux.Sensor{
			SensorTypeValue: linux.SensorIdleState,
			IsBinary:        true,
		},
	}

	var idleState bool

	var idleTime int64

	var err error

	if idleState, err = dbusx.GetProp[bool](ctx, dbusx.SystemBus, loginBasePath, loginBaseInterface, idleProp); err != nil {
		return nil, fmt.Errorf("could not retrieve idle state: %w", err)
	}

	newSensor.Value = idleState

	idleTime, err = dbusx.GetProp[int64](ctx, dbusx.SystemBus, loginBasePath, loginBaseInterface, idleTimeProp)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve idle time: %w", err)
	}

	newSensor.idleTime = idleTime

	return newSensor, nil
}

//nolint:cyclop,exhaustruct
//revive:disable:function-length
func IdleUpdater(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details)

	idleSensor, err := newIdleSensor(ctx)
	if err != nil {
		log.Debug().Err(err).Msg("Cannot create idle sensor.")
		close(sensorCh)

		return sensorCh
	}

	sessionPath, err := dbusx.GetSessionPath(ctx)
	if err != nil {
		log.Debug().Err(err).Msg("Cannot create idle sensor.")
		close(sensorCh)

		return sensorCh
	}

	events, err := dbusx.WatchBus(ctx, &dbusx.Watch{
		Bus:       dbusx.SystemBus,
		Names:     []string{sessionIdleProp, sessionIdleTimeProp},
		Interface: sessionInterface,
		Path:      sessionPath,
	})
	if err != nil {
		log.Debug().Err(err).
			Msg("Failed to create idle time D-Bus watch.")
		close(sensorCh)

		return sensorCh
	}

	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("Stopped idle state sensor.")

				return
			case event := <-events:
				if event.Signal == dbusx.PropChangedSignal {
					props, err := dbusx.ParsePropertiesChanged(event.Content)
					if err != nil {
						log.Warn().Err(err).Msg("Did not understand received trigger.")

						continue
					}

					if state, idleChanged := props.Changed[sessionIdleProp]; idleChanged {
						idleSensor.Value = dbusx.VariantToValue[bool](state)
						sensorCh <- idleSensor
					}

					if state, timeChanged := props.Changed[sessionIdleTimeProp]; timeChanged {
						idleSensor.idleTime = dbusx.VariantToValue[int64](state)
						sensorCh <- idleSensor
					}
				}
			}
		}
	}()

	// Send an initial sensor update.
	go func() {
		idleSensor, err := newIdleSensor(ctx)
		if err != nil {
			log.Debug().Err(err).Msg("Cannot start idle sensor.")

			return
		}
		sensorCh <- idleSensor
	}()

	return sensorCh
}

func idleTime(current int64) float64 {
	epoch := time.Unix(0, 0)
	uptime := time.Unix(current/idleSF, 0)

	return uptime.Sub(epoch).Seconds()
}
