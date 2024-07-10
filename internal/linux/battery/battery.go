// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:exhaustruct
//revive:disable:unused-receiver
package battery

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/iancoleman/strcase"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
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

	workerID = "battery_sensors"
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
	logger   *slog.Logger
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
			b.logger.Warn("Could not retrieve battery sensor.", "sensor", batterySensor.String(), "error", err.Error())

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
func newBattery(ctx context.Context, logger *slog.Logger, path dbus.ObjectPath) (*upowerBattery, error) {
	battery := &upowerBattery{
		dBusPath: path,
	}

	var (
		variant dbus.Variant
		err     error
	)

	// Get the battery type. Depending on the value, additional sensors will be added.
	variant, err = battery.getProp(ctx, linux.SensorBattType)
	if err != nil {
		return nil, fmt.Errorf("could not determine battery type: %w", err)
	}
	// Store the battery type.
	battery.battType, err = dbusx.VariantToValue[batteryType](variant)
	if err != nil {
		return nil, fmt.Errorf("could not determine battery type: %w", err)
	}

	// Use the native path D-Bus property for the battery id.
	variant, err = battery.getProp(ctx, linux.SensorBattNativePath)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve battery path in D-Bus: %w", err)
	}
	// Store the battery id/name.
	battery.id, err = dbusx.VariantToValue[string](variant)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve battery path in D-Bus: %w", err)
	}

	// Set up a logger for the battery with some battery-specific default
	// attributes.
	battery.logger = logger.With(
		slog.Group("battery_info",
			slog.String("name", battery.id),
			slog.String("dbus_path", string(battery.dBusPath)),
		),
	)

	// Get the battery model.
	variant, err = battery.getProp(ctx, linux.SensorBattModel)
	if err != nil {
		battery.logger.Warn("Could not determine battery model.")
	}
	// Store the battery model.
	battery.model, err = dbusx.VariantToValue[string](variant)
	if err != nil {
		battery.logger.Warn("Could not determine battery model.")
	}

	// At a minimum, monitor the battery type and the charging state.
	battery.sensors = append(battery.sensors, linux.SensorBattState)

	if battery.battType == 2 {
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
	logger     *slog.Logger
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

func (s *upowerBatterySensor) Attributes() map[string]any {
	attributes := make(map[string]any)

	attributes["extra_attributes"] = s.attributes

	return attributes
}

//nolint:exhaustive
func (s *upowerBatterySensor) generateAttributes(ctx context.Context, battery *upowerBattery) {
	switch s.SensorTypeValue {
	case linux.SensorBattEnergyRate:
		var (
			variant         dbus.Variant
			err             error
			voltage, energy float64
		)

		variant, err = battery.getProp(ctx, linux.SensorBattVoltage)
		if err != nil {
			s.logger.Warn("Could not retrieve battery voltage.", "error", err.Error())
		}

		voltage, err = dbusx.VariantToValue[float64](variant)
		if err != nil {
			s.logger.Warn("Could not retrieve battery voltage.", "error", err.Error())
		}

		variant, err = battery.getProp(ctx, linux.SensorBattEnergy)
		if err != nil {
			s.logger.Warn("Could not retrieve battery energy.", "error", err.Error())
		}

		energy, err = dbusx.VariantToValue[float64](variant)
		if err != nil {
			s.logger.Warn("Could not retrieve battery energy.", "error", err.Error())
		}

		s.attributes = &struct {
			DataSource string  `json:"data_source"`
			Voltage    float64 `json:"voltage"`
			Energy     float64 `json:"energy"`
		}{
			Voltage:    voltage,
			Energy:     energy,
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
		logger:    battery.logger,
	}
	batterySensor.SensorTypeValue = sensorType
	batterySensor.Value = value.Value()
	batterySensor.IsDiagnostic = true
	batterySensor.generateAttributes(ctx, battery)

	return batterySensor
}

type batteryTracker struct {
	batteryList map[dbus.ObjectPath]context.CancelFunc
	logger      *slog.Logger
	mu          sync.Mutex
}

func (t *batteryTracker) track(ctx context.Context, batteryPath dbus.ObjectPath) <-chan sensor.Details {
	battery, err := newBattery(ctx, t.logger, batteryPath)
	if err != nil {
		t.logger.Warn("Cannot monitor battery.", "path", batteryPath, "error", err.Error())

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
func newBatteryTracker(logger *slog.Logger) *batteryTracker {
	return &batteryTracker{
		batteryList: make(map[dbus.ObjectPath]context.CancelFunc),
		logger:      logger.With(slog.String("source", "battery_tracker")),
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
		battery.logger.Debug("Failed to create D-Bus watch for battery property changes.", "error", err.Error())
		close(sensorCh)

		return sensorCh
	}

	go func() {
		battery.logger.Debug("Monitoring battery.")

		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				battery.logger.Debug("Stopped monitoring battery.")

				return
			case event := <-events:
				props, err := dbusx.ParsePropertiesChanged(event.Content)
				if err != nil {
					battery.logger.Warn("Received a battery property change event that could not be understood.", "error", err.Error())

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
		tracker.logger.Debug("Failed to create D-Bus watch for battery additions/removals.", "error", err.Error())
		close(sensorCh)

		return sensorCh
	}

	go func() {
		tracker.logger.Debug("Monitoring for battery additions/removals.")

		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				tracker.logger.Debug("Stopped monitoring for batteries.")

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

type batterySensorWorker struct {
	logger *slog.Logger
}

// ?: implement initial battery sensor retrieval.
func (w *batterySensorWorker) Sensors(_ context.Context) ([]sensor.Details, error) {
	return nil, linux.ErrUnimplemented
}

//nolint:prealloc
func (w *batterySensorWorker) Events(ctx context.Context) (chan sensor.Details, error) {
	batteryTracker := newBatteryTracker(w.logger)

	var sensorCh []<-chan sensor.Details

	// Get a list of all current connected batteries and monitor them.
	batteries, err := getBatteries(ctx)
	if err != nil {
		w.logger.Warn("Could not retrieve any battery details from D-Bus.", "error", err.Error())
	}

	for _, path := range batteries {
		sensorCh = append(sensorCh, batteryTracker.track(ctx, path))
	}

	// Monitor for battery added/removed signals.
	sensorCh = append(sensorCh, monitorBatteryChanges(ctx, batteryTracker))

	return sensor.MergeSensorCh(ctx, sensorCh...), nil
}

func NewBatteryWorker(ctx context.Context) (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			Value: &batterySensorWorker{
				logger: logging.FromContext(ctx).With(slog.String("worker", workerID)),
			},
			WorkerID: workerID,
		},
		nil
}
