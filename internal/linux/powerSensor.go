// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/iancoleman/strcase"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/rs/zerolog/log"
)

const (
	powerProfilesDBusPath = "/net/hadess/PowerProfiles"
	powerProfilesDBusDest = "net.hadess.PowerProfiles"
)

type powerSensor struct {
	sensorValue      interface{}
	sensorAttributes interface{}
	sensorGroup      string
	sensorType       sensorType
}

func (state *powerSensor) Name() string {
	return state.sensorType.String()
}

func (state *powerSensor) ID() string {
	return strcase.ToSnake(state.sensorType.String())
}

func (state *powerSensor) Icon() string {
	return "mdi:flash"
}

func (state *powerSensor) SensorType() sensor.SensorType {
	return sensor.TypeSensor
}

func (state *powerSensor) DeviceClass() sensor.SensorDeviceClass {
	return 0
}

func (state *powerSensor) StateClass() sensor.SensorStateClass {
	return 0
}

func (state *powerSensor) State() interface{} {
	return state.sensorValue
}

func (state *powerSensor) Units() string {
	return ""
}

func (state *powerSensor) Category() string {
	return "diagnostic"
}

func (state *powerSensor) Attributes() interface{} {
	return state.sensorAttributes
}

func marshalPowerStateUpdate(sensor sensorType, group string, v dbus.Variant) *powerSensor {
	var value, attributes interface{}
	switch sensor {
	case powerProfile:
		value = strings.Trim(v.String(), "\"")
	}
	if attributes == nil {
		attributes = struct {
			DataSource string `json:"Data Source"`
		}{
			DataSource: "D-Bus",
		}
	}
	return &powerSensor{
		sensorGroup:      group,
		sensorType:       sensor,
		sensorValue:      value,
		sensorAttributes: attributes,
	}
}

func PowerUpater(ctx context.Context, status chan interface{}) {
	activePowerProfile, err := NewBusRequest(SystemBus).
		Path(powerProfilesDBusPath).
		Destination(powerProfilesDBusDest).
		GetProp(powerProfilesDBusDest + ".ActiveProfile")
	if err != nil {
		log.Debug().Err(err).Msg("Cannot retrieve a power profile from DBus.")
		return
	}

	status <- marshalPowerStateUpdate(powerProfile, powerProfilesDBusPath, activePowerProfile)

	err = NewBusRequest(SystemBus).
		Path(powerProfilesDBusPath).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(powerProfilesDBusPath),
		}).
		Event("org.freedesktop.DBus.Properties.PropertiesChanged").
		Handler(func(s *dbus.Signal) {
			updatedProps := s.Body[1].(map[string]dbus.Variant)
			for propName, propValue := range updatedProps {
				var propType sensorType
				switch propName {
				case "ActiveProfile":
					propType = powerProfile
				default:
					log.Debug().Msgf("Unhandled property %v changed to %v", propName, propValue)
				}
				if propType != 0 {
					propState := marshalPowerStateUpdate(propType, string(s.Path), propValue)
					status <- propState
				}
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Failed to create power state DBus watch.")
	}
}
