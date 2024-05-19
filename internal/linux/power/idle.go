// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

// revive:disable:unused-receiver

package power

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	idleIcon     = "mdi:sleep"
	notIdleIcon  = "mdi:sleep-off"
	idleProp     = managerInterface + ".IdleHint"
	idleTimeProp = managerInterface + ".IdleSinceHint"
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
		DataSource string  `json:"Data Source"`
		Seconds    float64 `json:"Duration"`
	}{
		DataSource: linux.DataSrcDbus,
		Seconds:    idleTime(s.idleTime),
	}
}

func IdleUpdater(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details)
	currentState := &idleSensor{}
	currentState.SensorTypeValue = linux.SensorIdleState

	err := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(loginBasePath),
			dbus.WithMatchInterface(managerInterface),
		}).
		Handler(func(s *dbus.Signal) {
			if !strings.Contains(string(s.Path), loginBasePath) || len(s.Body) <= 1 {
				return
			}
			if s.Name == dbusx.PropChangedSignal {
				props, ok := s.Body[1].(map[string]dbus.Variant)
				if !ok {
					return
				}
				for k, v := range props {
					switch k {
					case idleProp:
						currentState.Value = dbusx.VariantToValue[bool](v)
						sensorCh <- currentState
					case idleTimeProp:
						currentState.idleTime = dbusx.VariantToValue[int64](v)
						sensorCh <- currentState
					}
				}
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Warn().Err(err).
			Msg("Unable to monitor power state.")
		close(sensorCh)
		return sensorCh
	}

	// Send an initial sensor update.
	go func(s *idleSensor) {
		var idleState bool
		var idleTime int64
		var err error
		req := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
			Path(loginBasePath).
			Destination(loginBaseInterface)
		if idleState, err = dbusx.GetProp[bool](req, idleProp); err != nil {
			log.Debug().Err(err).Str("prop", filepath.Ext(idleProp)).Msg("Could not retrieve property from D-Bus.")
			return
		}
		s.Value = idleState
		idleTime, _ = dbusx.GetProp[int64](req, idleProp)
		s.idleTime = idleTime
		sensorCh <- s
	}(currentState)

	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Debug().Msg("Stopped idle state sensor.")
	}()
	return sensorCh
}

func idleTime(current int64) float64 {
	epoch := time.Unix(0, 0)
	uptime := time.Unix(current/1000, 0)
	return uptime.Sub(epoch).Seconds()
}
