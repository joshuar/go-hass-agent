// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package battery

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/iancoleman/strcase"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	upowerDBusDest         = "org.freedesktop.UPower"
	upowerDBusDeviceDest   = upowerDBusDest + ".Device"
	upowerDBusPath         = "/org/freedesktop/UPower"
	upowerGetDevicesMethod = "org.freedesktop.UPower.EnumerateDevices"

	deviceAddedSignal   = "DeviceAdded"
	deviceRemovedSignal = "DeviceRemoved"

	batteryIcon = "mdi:battery"
)

var ErrInvalidBattery = errors.New("invalid battery")

// dBusSensorToProps is a map of battery sensors to their D-Bus properties.
var dBusSensorToProps = map[linux.SensorTypeValue]string{
	linux.SensorBattType:       upowerDBusDeviceDest + ".Type",
	linux.SensorBattPercentage: upowerDBusDeviceDest + ".Percentage",
	linux.SensorBattTemp:       upowerDBusDeviceDest + ".Temperature",
	linux.SensorBattVoltage:    upowerDBusDeviceDest + ".Voltage",
	linux.SensorBattEnergy:     upowerDBusDeviceDest + ".Energy",
	linux.SensorBattEnergyRate: upowerDBusDeviceDest + ".EnergyRate",
	linux.SensorBattState:      upowerDBusDeviceDest + ".State",
	linux.SensorBattNativePath: upowerDBusDeviceDest + ".NativePath",
	linux.SensorBattLevel:      upowerDBusDeviceDest + ".BatteryLevel",
	linux.SensorBattModel:      upowerDBusDeviceDest + ".Model",
}

// dBusPropToSensor provides a map for to convert D-Bus properties to sensors.
var dBusPropToSensor = map[string]linux.SensorTypeValue{
	"Energy":       linux.SensorBattEnergy,
	"EnergyRate":   linux.SensorBattEnergyRate,
	"Voltage":      linux.SensorBattVoltage,
	"Percentage":   linux.SensorBattPercentage,
	"Temperatute":  linux.SensorBattTemp,
	"State":        linux.SensorBattState,
	"BatteryLevel": linux.SensorBattLevel,
}

type upowerBattery struct {
	id       string
	model    string
	dBusPath dbus.ObjectPath
	sensors  []linux.SensorTypeValue
	battType batteryType
}

// getProp retrieves the property from D-Bus that matches the given battery sensor type.
func (b *upowerBattery) getProp(ctx context.Context, t linux.SensorTypeValue) (dbus.Variant, error) {
	if !b.dBusPath.IsValid() {
		return dbus.MakeVariant(""), ErrInvalidBattery
	}

	return dbusx.GetProp[dbus.Variant](ctx, dbusx.SystemBus, string(b.dBusPath), upowerDBusDest, dBusSensorToProps[t])
}

// getSensors retrieves the sensors passed in for a given battery.
func (b *upowerBattery) getSensors(ctx context.Context, sensors ...linux.SensorTypeValue) chan sensor.Details {
	sensorCh := make(chan sensor.Details, len(sensors))
	defer close(sensorCh)

	for _, batterySensor := range sensors {
		value, err := b.getProp(ctx, batterySensor)
		if err != nil {
			log.Warn().Err(err).Str("battery", string(b.dBusPath)).Str("sensor", batterySensor.String()).Msg("Could not retrieve battery sensor.")

			continue
		}
		sensorCh <- newBatterySensor(ctx, b, batterySensor, value)
	}

	return sensorCh
}

// newBattery creates a battery object that will have a number of properties to
// be treated as sensors in Home Assistant.
//
//nolint:exhaustruct,mnd
func newBattery(ctx context.Context, path dbus.ObjectPath) (*upowerBattery, error) {
	battery := &upowerBattery{
		dBusPath: path,
	}

	// Get the battery type. Depending on the value, additional sensors will be added.
	battType, err := battery.getProp(ctx, linux.SensorBattType)
	if err != nil {
		return nil, fmt.Errorf("could not determine battery type: %w", err)
	}

	battery.battType = dbusx.VariantToValue[batteryType](battType)

	// use the native path D-Bus property for the battery id.
	id, err := battery.getProp(ctx, linux.SensorBattNativePath)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve battery path in D-Bus: %w", err)
	}

	battery.id = dbusx.VariantToValue[string](id)

	model, err := battery.getProp(ctx, linux.SensorBattModel)
	if err != nil {
		log.Warn().Err(err).Str("battery", string(battery.dBusPath)).Msg("Could not determine battery model.")
	}

	battery.model = dbusx.VariantToValue[string](model)

	// At a minimum, monitor the battery type and the charging state.
	battery.sensors = append(battery.sensors, linux.SensorBattState)

	if dbusx.VariantToValue[uint32](battType) == 2 {
		// Battery has charge percentage, temp and charging rate sensors
		battery.sensors = append(battery.sensors, linux.SensorBattPercentage, linux.SensorBattTemp, linux.SensorBattEnergyRate)
	} else {
		// Battery has a textual level sensor
		battery.sensors = append(battery.sensors, linux.SensorBattLevel)
	}

	return battery, nil
}

type upowerBatterySensor struct {
	attributes any
	batteryID  string
	model      string
	linux.Sensor
}

// uPowerBatteryState implements hass.SensorUpdate

func (s *upowerBatterySensor) Name() string {
	if s.model == "" {
		return s.batteryID + " " + s.SensorTypeValue.String()
	}

	return s.model + " " + s.SensorTypeValue.String()
}

func (s *upowerBatterySensor) ID() string {
	return s.batteryID + "_" + strings.ToLower(strcase.ToSnake(s.SensorTypeValue.String()))
}

//nolint:exhaustive
func (s *upowerBatterySensor) Icon() string {
	switch s.SensorTypeValue {
	case linux.SensorBattPercentage:
		return batteryPercentIcon(s.Value)
	case linux.SensorBattEnergyRate:
		return batteryChargeIcon(s.Value)
	default:
		return batteryIcon
	}
}

//nolint:exhaustive
func (s *upowerBatterySensor) DeviceClass() types.DeviceClass {
	switch s.SensorTypeValue {
	case linux.SensorBattPercentage:
		return types.DeviceClassBattery
	case linux.SensorBattTemp:
		return types.DeviceClassTemperature
	case linux.SensorBattEnergyRate:
		return types.DeviceClassPower
	default:
		return 0
	}
}

//nolint:exhaustive
func (s *upowerBatterySensor) StateClass() types.StateClass {
	switch s.SensorTypeValue {
	case linux.SensorBattPercentage, linux.SensorBattTemp, linux.SensorBattEnergyRate:
		return types.StateClassMeasurement
	default:
		return 0
	}
}

//nolint:exhaustive
func (s *upowerBatterySensor) State() any {
	if s.Value == nil {
		return sensor.StateUnknown
	}

	switch s.SensorTypeValue {
	case linux.SensorBattVoltage, linux.SensorBattTemp, linux.SensorBattEnergy, linux.SensorBattEnergyRate, linux.SensorBattPercentage:
		if value, ok := s.Value.(float64); !ok {
			return sensor.StateUnknown
		} else {
			return value
		}
	case linux.SensorBattState:
		if value, ok := s.Value.(uint32); !ok {
			return sensor.StateUnknown
		} else {
			return battChargeState(value).String()
		}
	case linux.SensorBattLevel:
		if value, ok := s.Value.(uint32); !ok {
			return sensor.StateUnknown
		} else {
			return batteryLevel(value).String()
		}
	default:
		if value, ok := s.Value.(string); !ok {
			return sensor.StateUnknown
		} else {
			return value
		}
	}
}

//nolint:exhaustive
func (s *upowerBatterySensor) Units() string {
	switch s.SensorTypeValue {
	case linux.SensorBattPercentage:
		return "%"
	case linux.SensorBattTemp:
		return "°C"
	case linux.SensorBattEnergyRate:
		return "W"
	default:
		return ""
	}
}

func (s *upowerBatterySensor) Attributes() any {
	return s.attributes
}

//nolint:exhaustive
func (s *upowerBatterySensor) generateAttributes(ctx context.Context, battery *upowerBattery) {
	switch s.SensorTypeValue {
	case linux.SensorBattEnergyRate:
		voltage, err := battery.getProp(ctx, linux.SensorBattVoltage)
		if err != nil {
			log.Warn().Err(err).Str("battery", string(battery.dBusPath)).Msg("Could not retrieve battery voltage.")
		}

		energy, err := battery.getProp(ctx, linux.SensorBattEnergy)
		if err != nil {
			log.Warn().Err(err).Str("battery", string(battery.dBusPath)).Msg("Could not retrieve battery energy.")
		}

		s.attributes = &struct {
			DataSource string  `json:"data_source"`
			Voltage    float64 `json:"voltage"`
			Energy     float64 `json:"energy"`
		}{
			Voltage:    dbusx.VariantToValue[float64](voltage),
			Energy:     dbusx.VariantToValue[float64](energy),
			DataSource: linux.DataSrcDbus,
		}
	case linux.SensorBattPercentage, linux.SensorBattLevel:
		s.attributes = &struct {
			Type       string `json:"battery_type"`
			DataSource string `json:"data_source"`
		}{
			Type:       battery.battType.String(),
			DataSource: linux.DataSrcDbus,
		}
	}
}

// newBatterySensor creates a new sensor for Home Assistant from a battery
// property.
//
//nolint:exhaustruct,lll
func newBatterySensor(ctx context.Context, battery *upowerBattery, sensorType linux.SensorTypeValue, value dbus.Variant) *upowerBatterySensor {
	batterySensor := &upowerBatterySensor{
		batteryID: battery.id,
		model:     battery.model,
	}
	batterySensor.SensorTypeValue = sensorType
	batterySensor.Value = value.Value()
	batterySensor.IsDiagnostic = true
	batterySensor.generateAttributes(ctx, battery)

	return batterySensor
}

type batteryTracker struct {
	batteryList map[dbus.ObjectPath]context.CancelFunc
	mu          sync.Mutex
}

func (t *batteryTracker) track(ctx context.Context, batteryPath dbus.ObjectPath) <-chan sensor.Details {
	battery, err := newBattery(ctx, batteryPath)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot monitory battery.")

		return nil
	}

	battCtx, cancelFunc := context.WithCancel(ctx)

	t.mu.Lock()
	t.batteryList[batteryPath] = cancelFunc
	t.mu.Unlock()

	return sensor.MergeSensorCh(battCtx, battery.getSensors(battCtx, battery.sensors...), monitorBattery(battCtx, battery))
}

func (t *batteryTracker) remove(batteryPath dbus.ObjectPath) {
	if cancelFunc, ok := t.batteryList[batteryPath]; ok {
		cancelFunc()
		t.mu.Lock()
		delete(t.batteryList, batteryPath)
		t.mu.Unlock()
	}
}

//nolint:exhaustruct
func newBatteryTracker() *batteryTracker {
	return &batteryTracker{
		batteryList: make(map[dbus.ObjectPath]context.CancelFunc),
	}
}

// getBatteries is a helper function to retrieve all of the known batteries
// connected to the system.
func getBatteries(ctx context.Context) ([]dbus.ObjectPath, error) {
	batteryList, err := dbusx.GetData[[]dbus.ObjectPath](ctx, dbusx.SystemBus, upowerDBusPath, upowerDBusDest, upowerGetDevicesMethod)
	if err != nil {
		return nil, err
	}

	return batteryList, nil
}

// monitorBattery will monitor a battery device for any property changes and
// send these as sensors.
//
//nolint:exhaustruct
func monitorBattery(ctx context.Context, battery *upowerBattery) <-chan sensor.Details {
	log.Debug().Str("battery", battery.id).Msg("Monitoring battery.")

	sensorCh := make(chan sensor.Details)
	// Create a DBus signal match to watch for property changes for this
	// battery.
	events, err := dbusx.WatchBus(ctx, &dbusx.Watch{
		Bus:       dbusx.SystemBus,
		Names:     []string{dbusx.PropChangedSignal},
		Path:      string(battery.dBusPath),
		Interface: dbusx.PropInterface,
	})
	if err != nil {
		log.Debug().Err(err).
			Msg("Failed to create battery props D-Bus watch.")
		close(sensorCh)

		return sensorCh
	}

	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				log.Debug().Str("battery", battery.id).Msg("Stopped monitoring battery.")

				return
			case event := <-events:
				props, err := dbusx.ParsePropertiesChanged(event.Content)
				if err != nil {
					log.Warn().Err(err).Msg("Did not understand received trigger.")

					continue
				}

				for prop, value := range props.Changed {
					if s, ok := dBusPropToSensor[prop]; ok {
						sensorCh <- newBatterySensor(ctx, battery, s, value)
					}
				}
			}
		}
	}()

	return sensorCh
}

// monitorBatteryChanges monitors for battery devices being added/removed from
// the system and will start/stop monitory each battery as appropriate.
//
//nolint:exhaustruct
func monitorBatteryChanges(ctx context.Context, tracker *batteryTracker) <-chan sensor.Details {
	sensorCh := make(chan sensor.Details)

	events, err := dbusx.WatchBus(ctx, &dbusx.Watch{
		Bus:       dbusx.SystemBus,
		Names:     []string{deviceAddedSignal, deviceRemovedSignal},
		Interface: upowerDBusDest,
		Path:      upowerDBusPath,
	})
	if err != nil {
		log.Debug().Err(err).
			Msg("Failed to create battery state D-Bus watch.")
		close(sensorCh)

		return sensorCh
	}

	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("Stopped monitoring for batteries.")

				return
			case event := <-events:
				batteryPath, validBatteryPath := event.Content[0].(dbus.ObjectPath)
				if !validBatteryPath {
					continue
				}

				switch {
				case strings.Contains(event.Signal, deviceAddedSignal):
					go func() {
						for s := range tracker.track(ctx, batteryPath) {
							sensorCh <- s
						}
					}()
				case strings.Contains(event.Signal, deviceRemovedSignal):
					tracker.remove(batteryPath)
				}
			}
		}
	}()

	return sensorCh
}

//nolint:mnd
func batteryPercentIcon(v any) string {
	percentage, ok := v.(float64)
	if !ok {
		return batteryIcon + "-unknown"
	}

	if percentage >= 95 {
		return batteryIcon
	}

	return fmt.Sprintf("%s-%d", batteryIcon, int(math.Round(percentage/10)*10))
}

func batteryChargeIcon(v any) string {
	energyRate, ok := v.(float64)
	if !ok {
		return batteryIcon
	}

	if math.Signbit(energyRate) {
		return batteryIcon + "-minus"
	}

	return batteryIcon + "-plus"
}

type worker struct{}

// ?: implement initial battery sensor retrieval.
func (w *worker) Sensors(_ context.Context) ([]sensor.Details, error) {
	return nil, linux.ErrUnimplemented
}

//nolint:prealloc
func (w *worker) Events(ctx context.Context) chan sensor.Details {
	batteryTracker := newBatteryTracker()

	var sensorCh []<-chan sensor.Details

	// Get a list of all current connected batteries and monitor them.
	batteries, err := getBatteries(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Could not retrieve battery list. Cannot find any existing batteries.")
	}

	for _, path := range batteries {
		sensorCh = append(sensorCh, batteryTracker.track(ctx, path))
	}

	// Monitor for battery added/removed signals.
	sensorCh = append(sensorCh, monitorBatteryChanges(ctx, batteryTracker))

	return sensor.MergeSensorCh(ctx, sensorCh...)
}

func NewBatteryWorker(_ context.Context) (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			WorkerName: "Battery Sensors",
			WorkerDesc: "Sensors to track connected battery states.",
			Value:      &worker{},
		},
		nil
}
