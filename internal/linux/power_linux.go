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
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

//go:generate stringer -type=powerProp -output power_props_linux.go
const (
	powerProfilesDBusPath           = "/net/hadess/PowerProfiles"
	powerProfilesDBusDest           = "net.hadess.PowerProfiles"
	profile               powerProp = iota + 1
)

type powerProp int

type powerSensor struct {
	sensorGroup      string
	sensorType       powerProp
	sensorValue      interface{}
	sensorAttributes interface{}
}

func (state *powerSensor) Name() string {
	switch state.sensorType {
	case profile:
		return "Power Profile"
	default:
		return state.sensorGroup + " " + strcase.ToDelimited(state.sensorType.String(), ' ')
	}
}

func (state *powerSensor) ID() string {
	switch state.sensorType {
	case profile:
		return "power_profile"
	default:
		return state.sensorGroup + "_" + strcase.ToSnake(state.sensorType.String())
	}
}

func (state *powerSensor) Icon() string {
	return "mdi:flash"
}

func (state *powerSensor) SensorType() hass.SensorType {
	return hass.TypeSensor
}

func (state *powerSensor) DeviceClass() hass.SensorDeviceClass {
	return 0
}

func (state *powerSensor) StateClass() hass.SensorStateClass {
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

func marshalPowerStateUpdate(ctx context.Context, sensor powerProp, path dbus.ObjectPath, group string, v dbus.Variant) *powerSensor {
	var value, attributes interface{}
	switch sensor {
	case profile:
		value = strings.Trim(v.String(), "\"")
	}
	return &powerSensor{
		sensorGroup:      group,
		sensorType:       sensor,
		sensorValue:      value,
		sensorAttributes: attributes,
	}
}

func PowerUpater(ctx context.Context, status chan interface{}) {
	deviceAPI, err := FetchAPIFromContext(ctx)
	if err != nil {
		log.Debug().Err(err).Caller().
			Msg("Could not connect to DBus.")
		return
	}

	activePowerProfile, err := deviceAPI.SystemBusRequest().
		Path(powerProfilesDBusPath).
		Destination(powerProfilesDBusDest).
		GetProp(powerProfilesDBusDest + ".ActiveProfile")
	if err != nil {
		log.Debug().Err(err).Msg("Cannot retrieve a power profile from DBus.")
		return
	}

	status <- marshalPowerStateUpdate(ctx,
		profile,
		powerProfilesDBusPath,
		"",
		activePowerProfile)

	powerProfileDBusMatch := []dbus.MatchOption{
		dbus.WithMatchObjectPath(powerProfilesDBusPath),
	}
	powerProfileHandler := func(s *dbus.Signal) {
		updatedProps := s.Body[1].(map[string]dbus.Variant)
		for propName, propValue := range updatedProps {
			var propType powerProp
			switch propName {
			case "ActiveProfile":
				propType = profile
			default:
				log.Debug().Msgf("Unhandled property %v changed to %v", propName, propValue)
			}
			if propType != 0 {
				propState := marshalPowerStateUpdate(ctx,
					propType,
					s.Path,
					"",
					propValue)
				status <- propState
			}
		}
	}
	deviceAPI.SystemBusRequest().
		Path(powerProfilesDBusPath).
		Match(powerProfileDBusMatch).
		Event("org.freedesktop.DBus.Properties.PropertiesChanged").
		Handler(powerProfileHandler).
		AddWatch()
}
