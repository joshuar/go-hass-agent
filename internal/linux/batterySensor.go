// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/iancoleman/strcase"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/tracker"
	"github.com/joshuar/go-hass-agent/pkg/dbushelpers"
	"github.com/rs/zerolog/log"
)

const (
	upowerDBusDest         = "org.freedesktop.UPower"
	upowerDBusDeviceDest   = upowerDBusDest + ".Device"
	upowerDBusPath         = "/org/freedesktop/UPower"
	upowerGetDevicesMethod = "org.freedesktop.UPower.EnumerateDevices"
)

type upowerBattery struct {
	id, model string
	dBusPath  dbus.ObjectPath
	battType  batteryType
	sensors   []sensorType
}

// getProp retrieves the property from D-Bus that matches the given battery sensor type.
func (b *upowerBattery) getProp(ctx context.Context, t sensorType) (dbus.Variant, error) {
	if !b.dBusPath.IsValid() {
		return dbus.MakeVariant(""), errors.New("invalid battery path")
	}
	dBusProp := sensorTypeToDBusProp(t)
	return dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Path(b.dBusPath).
		Destination(upowerDBusDest).
		GetProp(dBusProp)
}

// getSensors retrieves the sensors passed in for a given battery.
func (b *upowerBattery) getSensors(ctx context.Context, sensors ...sensorType) chan *upowerBatterySensor {
	sensorCh := make(chan *upowerBatterySensor, len(sensors))
	for _, s := range sensors {
		value, err := b.getProp(ctx, s)
		if err != nil {
			log.Warn().Err(err).Str("battery", string(b.dBusPath)).Str("sensor", s.String()).Msg("Could not retrieve battery sensor.")
			continue
		}
		sensorCh <- newBatterySensor(ctx, b, s, value)
	}
	close(sensorCh)
	return sensorCh
}

func newBattery(ctx context.Context, path dbus.ObjectPath) *upowerBattery {
	b := &upowerBattery{
		dBusPath: path,
	}

	// Get the battery type. Depending on the value, additional sensors will be added.
	battType, err := b.getProp(ctx, battType)
	if err != nil {
		log.Warn().Err(err).Msg("Could not determine battery type.")
		return nil
	}
	b.battType = dbushelpers.VariantToValue[batteryType](battType)

	// use the native path D-Bus property for the battery id.
	id, err := b.getProp(ctx, battNativePath)
	if err != nil {
		log.Warn().Err(err).Str("battery", string(b.dBusPath)).Msg("Battery does not have a usable path. Can not monitor sensors.")
		return nil
	}
	b.id = dbushelpers.VariantToValue[string](id)

	model, err := b.getProp(ctx, battModel)
	if err != nil {
		log.Warn().Err(err).Str("battery", string(b.dBusPath)).Msg("Could not determine battery model.")
	}
	b.model = dbushelpers.VariantToValue[string](model)

	// At a minimum, monitor the battery type and the charging state.
	b.sensors = append(b.sensors, battState)

	if dbushelpers.VariantToValue[uint32](battType) == 2 {
		// Battery has charge percentage, temp and charging rate sensors
		b.sensors = append(b.sensors, battPercentage, battTemp, battEnergyRate)
	} else {
		// Battery has a textual level sensor
		b.sensors = append(b.sensors, battLevel, battPercentage)
	}
	return b
}

type upowerBatterySensor struct {
	attributes any
	batteryID  string
	model      string
	linuxSensor
}

// uPowerBatteryState implements hass.SensorUpdate

func (state *upowerBatterySensor) Name() string {
	return state.model + " " + state.sensorType.String()
}

func (state *upowerBatterySensor) ID() string {
	return state.batteryID + "_" + strings.ToLower(strcase.ToSnake(state.sensorType.String()))
}

func (state *upowerBatterySensor) Icon() string {
	switch state.sensorType {
	case battPercentage:
		return battPcToIcon(state.value)
	case battEnergyRate:
		return battErToIcon(state.value)
	default:
		return "mdi:battery"
	}
}

func (state *upowerBatterySensor) DeviceClass() sensor.SensorDeviceClass {
	switch state.sensorType {
	case battPercentage:
		return sensor.SensorBattery
	case battTemp:
		return sensor.SensorTemperature
	case battEnergyRate:
		return sensor.SensorPower
	default:
		return 0
	}
}

func (state *upowerBatterySensor) StateClass() sensor.SensorStateClass {
	switch state.sensorType {
	case battPercentage, battTemp, battEnergyRate:
		return sensor.StateMeasurement
	default:
		return 0
	}
}

func (state *upowerBatterySensor) State() any {
	if state.value == nil {
		return sensor.StateUnknown
	}
	switch state.sensorType {
	case battVoltage, battTemp, battEnergy, battEnergyRate, battPercentage:
		if value, ok := state.value.(float64); !ok {
			return sensor.StateUnknown
		} else {
			return value
		}
	case battState:
		if value, ok := state.value.(battChargeState); !ok {
			return sensor.StateUnknown
		} else {
			return value.String()
		}
	case battLevel:
		if value, ok := state.value.(batteryLevel); !ok {
			return sensor.StateUnknown
		} else {
			return value.String()
		}
	default:
		if value, ok := state.value.(string); !ok {
			return sensor.StateUnknown
		} else {
			return value
		}
	}
}

func (state *upowerBatterySensor) Units() string {
	switch state.sensorType {
	case battPercentage:
		return "%"
	case battTemp:
		return "°C"
	case battEnergyRate:
		return "W"
	default:
		return ""
	}
}

func (state *upowerBatterySensor) Attributes() any {
	return state.attributes
}

func (state *upowerBatterySensor) generateAttributes(ctx context.Context, b *upowerBattery) {
	switch state.sensorType {
	case battEnergyRate:
		voltage, err := b.getProp(ctx, battVoltage)
		if err != nil {
			log.Warn().Err(err).Str("battery", string(b.dBusPath)).Msg("Could not retrieve battery voltage.")
		}
		energy, err := b.getProp(ctx, battEnergy)
		if err != nil {
			log.Warn().Err(err).Str("battery", string(b.dBusPath)).Msg("Could not retrieve battery energy.")
		}
		state.attributes = &struct {
			DataSource string  `json:"Data Source"`
			Voltage    float64 `json:"Voltage"`
			Energy     float64 `json:"Energy"`
		}{
			Voltage:    dbushelpers.VariantToValue[float64](voltage),
			Energy:     dbushelpers.VariantToValue[float64](energy),
			DataSource: srcDbus,
		}
	case battPercentage, battLevel:
		state.attributes = &struct {
			Type       string `json:"Battery Type"`
			DataSource string `json:"Data Source"`
		}{
			Type:       b.battType.String(),
			DataSource: srcDbus,
		}
	}
}

func newBatterySensor(ctx context.Context, b *upowerBattery, t sensorType, v dbus.Variant) *upowerBatterySensor {
	s := &upowerBatterySensor{
		batteryID: b.id,
		model:     b.model,
	}
	s.sensorType = t
	s.value = v.Value()
	s.isDiagnostic = true
	s.generateAttributes(ctx, b)
	return s
}

func BatteryUpdater(ctx context.Context) chan tracker.Sensor {
	// D-Bus uses different names for the battery sensors we send to Home
	// Assistant. This map allows conversion between the two.
	dBusPropToSensor := map[string]sensorType{
		"Energy":       battEnergy,
		"EnergyRate":   battEnergyRate,
		"Voltage":      battVoltage,
		"Percentage":   battPercentage,
		"Temperatute":  battTemp,
		"State":        battState,
		"BatteryLevel": battLevel,
	}

	sensorCh := make(chan tracker.Sensor, 1)
	batteries := getBatteries(ctx)
	if len(batteries) < 1 {
		log.Warn().
			Msg("Unable to get any battery devices from D-Bus. Battery sensor will not run.")
		close(sensorCh)
		return sensorCh
	}

	for _, path := range batteries {
		// Track this battery in batteryTracker.
		battery := newBattery(ctx, path)

		// Send its current state as sensors.
		go func() {
			for s := range battery.getSensors(ctx, battery.sensors...) {
				sensorCh <- s
			}
		}()

		// Create a DBus signal match to watch for property changes for this
		// battery. If a property changes, check it is one we want to track and
		// if so, update the battery's state in batteryTracker and send the
		// update back to Home Assistant.
		err := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
			Path(path).
			Match([]dbus.MatchOption{
				dbus.WithMatchObjectPath(path),
				dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
			}).
			Event("org.freedesktop.DBus.Properties.PropertiesChanged").
			Handler(func(s *dbus.Signal) {
				if s.Path != path {
					return
				}
				if len(s.Body) == 0 {
					log.Debug().Msg("Received battery state change signal but not properties sent.")
					return
				}
				props, ok := s.Body[1].(map[string]dbus.Variant)
				if !ok {
					log.Debug().Msg("Could not map received signal to battery properties.")
					return
				}
				for propName, propValue := range props {
					if s, ok := dBusPropToSensor[propName]; ok {
						sensorCh <- newBatterySensor(ctx, battery, s, propValue)
					}
				}
			}).
			AddWatch(ctx)
		if err != nil {
			log.Debug().Caller().Err(err).
				Msg("Failed to create DBus battery property watch.")
			close(sensorCh)
			return sensorCh
		}
	}
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Debug().Msg("Stopped battery sensors.")
	}()
	return sensorCh
}

// getBatteries is a helper function to retrieve all of the known batteries
// connected to the system.
func getBatteries(ctx context.Context) []dbus.ObjectPath {
	return dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Path(upowerDBusPath).
		Destination(upowerDBusDest).
		GetData(upowerGetDevicesMethod).AsObjectPathList()
}

// sensorTypeToDBusProp is a helper function to generate the correct D-Bus
// property for the given sensor.
func sensorTypeToDBusProp(t sensorType) string {
	switch t {
	case battType:
		return upowerDBusDeviceDest + ".Type"
	case battPercentage:
		return upowerDBusDeviceDest + ".Percentage"
	case battTemp:
		return upowerDBusDeviceDest + ".Temperature"
	case battVoltage:
		return upowerDBusDeviceDest + ".Voltage"
	case battEnergy:
		return upowerDBusDeviceDest + ".Energy"
	case battEnergyRate:
		return upowerDBusDeviceDest + ".EnergyRate"
	case battState:
		return upowerDBusDeviceDest + ".State"
	case battNativePath:
		return upowerDBusDeviceDest + ".NativePath"
	case battLevel:
		return upowerDBusDeviceDest + ".BatteryLevel"
	case battModel:
		return upowerDBusDeviceDest + ".Model"
	}
	return ""
}

func battPcToIcon(v any) string {
	pc, ok := v.(float64)
	if !ok {
		return "mdi:battery-unknown"
	}
	if pc >= 95 {
		return "mdi:battery"
	}
	return fmt.Sprintf("mdi:battery-%d", int(math.Round(pc/10)*10))
}

func battErToIcon(v any) string {
	er, ok := v.(float64)
	if !ok {
		return "mdi:battery"
	}
	if math.Signbit(er) {
		return "mdi:battery-minus"
	}
	return "mdi:battery-plus"
}
