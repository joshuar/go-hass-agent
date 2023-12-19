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
	"sync"

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

// dBusSensorToProps is a map of battery sensors to their D-Bus properties.
var dBusSensorToProps = map[sensorType]string{
	battType:       upowerDBusDeviceDest + ".Type",
	battPercentage: upowerDBusDeviceDest + ".Percentage",
	battTemp:       upowerDBusDeviceDest + ".Temperature",
	battVoltage:    upowerDBusDeviceDest + ".Voltage",
	battEnergy:     upowerDBusDeviceDest + ".Energy",
	battEnergyRate: upowerDBusDeviceDest + ".EnergyRate",
	battState:      upowerDBusDeviceDest + ".State",
	battNativePath: upowerDBusDeviceDest + ".NativePath",
	battLevel:      upowerDBusDeviceDest + ".BatteryLevel",
	battModel:      upowerDBusDeviceDest + ".Model",
}

// dBusPropToSensor provides a map for to convert D-Bus properties to sensors.
var dBusPropToSensor = map[string]sensorType{
	"Energy":       battEnergy,
	"EnergyRate":   battEnergyRate,
	"Voltage":      battVoltage,
	"Percentage":   battPercentage,
	"Temperatute":  battTemp,
	"State":        battState,
	"BatteryLevel": battLevel,
}

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
	return dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Path(b.dBusPath).
		Destination(upowerDBusDest).
		GetProp(dBusSensorToProps[t])
}

// getSensors retrieves the sensors passed in for a given battery.
func (b *upowerBattery) getSensors(ctx context.Context, sensors ...sensorType) chan tracker.Sensor {
	sensorCh := make(chan tracker.Sensor, len(sensors))
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

// newBattery creates a battery object that will have a number of properties to
// be treated as sensors in Home Assistant.
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
		b.sensors = append(b.sensors, battLevel)
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

func (s *upowerBatterySensor) Name() string {
	if s.model == "" {
		return s.batteryID + " " + s.sensorType.String()
	}
	return s.model + " " + s.sensorType.String()
}

func (s *upowerBatterySensor) ID() string {
	return s.batteryID + "_" + strings.ToLower(strcase.ToSnake(s.sensorType.String()))
}

func (s *upowerBatterySensor) Icon() string {
	switch s.sensorType {
	case battPercentage:
		return battPcToIcon(s.value)
	case battEnergyRate:
		return battErToIcon(s.value)
	default:
		return "mdi:battery"
	}
}

func (s *upowerBatterySensor) DeviceClass() sensor.SensorDeviceClass {
	switch s.sensorType {
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

func (s *upowerBatterySensor) StateClass() sensor.SensorStateClass {
	switch s.sensorType {
	case battPercentage, battTemp, battEnergyRate:
		return sensor.StateMeasurement
	default:
		return 0
	}
}

func (s *upowerBatterySensor) State() any {
	if s.value == nil {
		return sensor.StateUnknown
	}
	switch s.sensorType {
	case battVoltage, battTemp, battEnergy, battEnergyRate, battPercentage:
		if value, ok := s.value.(float64); !ok {
			return sensor.StateUnknown
		} else {
			return value
		}
	case battState:
		if value, ok := s.value.(uint32); !ok {
			return sensor.StateUnknown
		} else {
			return battChargeState(value).String()
		}
	case battLevel:
		if value, ok := s.value.(uint32); !ok {
			return sensor.StateUnknown
		} else {
			return batteryLevel(value).String()
		}
	default:
		if value, ok := s.value.(string); !ok {
			return sensor.StateUnknown
		} else {
			return value
		}
	}
}

func (s *upowerBatterySensor) Units() string {
	switch s.sensorType {
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

func (s *upowerBatterySensor) Attributes() any {
	return s.attributes
}

func (s *upowerBatterySensor) generateAttributes(ctx context.Context, b *upowerBattery) {
	switch s.sensorType {
	case battEnergyRate:
		voltage, err := b.getProp(ctx, battVoltage)
		if err != nil {
			log.Warn().Err(err).Str("battery", string(b.dBusPath)).Msg("Could not retrieve battery voltage.")
		}
		energy, err := b.getProp(ctx, battEnergy)
		if err != nil {
			log.Warn().Err(err).Str("battery", string(b.dBusPath)).Msg("Could not retrieve battery energy.")
		}
		s.attributes = &struct {
			DataSource string  `json:"Data Source"`
			Voltage    float64 `json:"Voltage"`
			Energy     float64 `json:"Energy"`
		}{
			Voltage:    dbushelpers.VariantToValue[float64](voltage),
			Energy:     dbushelpers.VariantToValue[float64](energy),
			DataSource: srcDbus,
		}
	case battPercentage, battLevel:
		s.attributes = &struct {
			Type       string `json:"Battery Type"`
			DataSource string `json:"Data Source"`
		}{
			Type:       b.battType.String(),
			DataSource: srcDbus,
		}
	}
}

// newBatterySensor creates a new sensor for Home Assistant from a battery
// property.
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

type batteryTracker struct {
	batteryList map[dbus.ObjectPath]context.CancelFunc
	mu          sync.Mutex
}

func (t *batteryTracker) track(ctx context.Context, p dbus.ObjectPath) <-chan tracker.Sensor {
	battCtx, cancelFunc := context.WithCancel(ctx)
	t.mu.Lock()
	t.batteryList[p] = cancelFunc
	t.mu.Unlock()
	battery := newBattery(ctx, p)
	return tracker.MergeSensorCh(battCtx, battery.getSensors(battCtx, battery.sensors...), monitorBattery(battCtx, battery))
}

func (t *batteryTracker) remove(p dbus.ObjectPath) {
	if cancelFunc, ok := t.batteryList[p]; ok {
		cancelFunc()
		t.mu.Lock()
		delete(t.batteryList, p)
		t.mu.Unlock()
	}
}

func newBatteryTracker() *batteryTracker {
	return &batteryTracker{
		batteryList: make(map[dbus.ObjectPath]context.CancelFunc),
	}
}

func BatteryUpdater(ctx context.Context) chan tracker.Sensor {
	batteryTracker := newBatteryTracker()
	var sensorCh []<-chan tracker.Sensor

	// Get a list of all current connected batteries and monitor them.
	batteries := getBatteries(ctx)
	if len(batteries) > 0 {
		for _, path := range batteries {
			sensorCh = append(sensorCh, batteryTracker.track(ctx, path))
		}
	}

	// Monitor for battery added/removed signals.
	sensorCh = append(sensorCh, monitorBatteryChanges(ctx, batteryTracker))

	return tracker.MergeSensorCh(ctx, sensorCh...)
}

// getBatteries is a helper function to retrieve all of the known batteries
// connected to the system.
func getBatteries(ctx context.Context) []dbus.ObjectPath {
	return dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Path(upowerDBusPath).
		Destination(upowerDBusDest).
		GetData(upowerGetDevicesMethod).AsObjectPathList()
}

// monitorBattery will monitor a battery device for any property changes and
// send these as sensors.
func monitorBattery(ctx context.Context, battery *upowerBattery) <-chan tracker.Sensor {
	log.Debug().Str("battery", battery.id).
		Msg("Monitoring battery.")
	sensorCh := make(chan tracker.Sensor, 1)
	// Create a DBus signal match to watch for property changes for this
	// battery.
	err := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(battery.dBusPath),
			dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		}).
		Event(dbushelpers.PropChangedSignal).
		Handler(func(s *dbus.Signal) {
			if s.Path != battery.dBusPath || len(s.Body) == 0 {
				log.Trace().Caller().Msg("Not my signal or empty signal body.")
				return
			}
			props, ok := s.Body[1].(map[string]dbus.Variant)
			if !ok {
				return
			}
			go func() {
				for propName, propValue := range props {
					if s, ok := dBusPropToSensor[propName]; ok {
						sensorCh <- newBatterySensor(ctx, battery, s, propValue)
					}
				}
			}()
		}).
		AddWatch(ctx)
	if err != nil {
		log.Debug().Err(err).Str("battery", battery.id).
			Msg("Could not monitor D-Bus for battery properties.")
		close(sensorCh)
		return sensorCh
	}
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Debug().Str("battery", battery.id).
			Msg("Stopped monitoring battery.")
	}()
	return sensorCh
}

// monitorBatteryChanges monitors for battery devices being added/removed from
// the system and will start/stop monitory each battery as appropriate.
func monitorBatteryChanges(ctx context.Context, t *batteryTracker) <-chan tracker.Sensor {
	sensorCh := make(chan tracker.Sensor, 1)
	err := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(upowerDBusPath),
			dbus.WithMatchInterface(upowerDBusDest),
		}).
		Handler(func(s *dbus.Signal) {
			if !strings.Contains(s.Name, upowerDBusDest) {
				log.Trace().Caller().Msg("Not my signal.")
				return
			}
			var batteryPath dbus.ObjectPath
			var ok bool
			if batteryPath, ok = s.Body[0].(dbus.ObjectPath); !ok {
				return
			}
			switch s.Name {
			case "org.freedesktop.UPower.DeviceAdded":
				go func() {
					for s := range t.track(ctx, batteryPath) {
						sensorCh <- s
					}
				}()
			case "org.freedesktop.UPower.DeviceRemoved":
				t.remove(batteryPath)
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Failed to create DBus battery property watch.")
		close(sensorCh)
		return sensorCh
	}
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
	}()
	return sensorCh
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
