// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/host"
)

//go:generate stringer -type=timeProp -output timeSensorProps.go -linecomment
const (
	boottime timeProp = iota + 1 // Last Reboot
	uptime                       // Uptime
)

type timeProp int

type timeSensor struct {
	value interface{}
	prop  timeProp
}

func (m *timeSensor) Name() string {
	return m.prop.String()
}

func (m *timeSensor) ID() string {
	return strcase.ToSnake(m.prop.String())
}

func (m *timeSensor) Icon() string {
	return "mdi:restart"
}

func (m *timeSensor) SensorType() hass.SensorType {
	return hass.TypeSensor
}

func (m *timeSensor) DeviceClass() hass.SensorDeviceClass {
	switch m.prop {
	case uptime:
		return hass.Duration
	case boottime:
		return hass.Timestamp
	}
	return 0
}

func (m *timeSensor) StateClass() hass.SensorStateClass {
	switch m.prop {
	case uptime:
		return hass.StateMeasurement
	default:
		return 0
	}
}

func (m *timeSensor) State() interface{} {
	return m.value
}

func (m *timeSensor) Units() string {
	switch m.prop {
	case uptime:
		return "h"
	default:
		return ""
	}
}

func (m *timeSensor) Category() string {
	return "diagnostic"
}

func (m *timeSensor) Attributes() interface{} {
	switch m.prop {
	case uptime:
		return struct {
			NativeUnit string `json:"native_unit_of_measurement"`
		}{
			NativeUnit: "h",
		}
	default:
		return nil
	}
}

func TimeUpdater(ctx context.Context, status chan interface{}) {
	updateTimes := func() {
		status <- &timeSensor{
			prop:  uptime,
			value: getUptime(ctx),
		}

		status <- &timeSensor{
			prop:  boottime,
			value: getBoottime(ctx),
		}
	}

	pollSensors(ctx, updateTimes, time.Minute*15, time.Minute)
}

func getUptime(ctx context.Context) interface{} {
	u, err := host.UptimeWithContext(ctx)
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Failed to retrieve uptime.")
		return "Unknown"
	}
	epoch := time.Unix(0, 0)
	uptime := time.Unix(int64(u), 0)
	return uptime.Sub(epoch).Hours()
}

func getBoottime(ctx context.Context) string {
	u, err := host.BootTimeWithContext(ctx)
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Failed to retrieve boottime.")
		return "Unknown"
	}
	return time.Unix(int64(u), 0).Format(time.RFC3339)
}
