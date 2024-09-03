// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
//go:generate stringer -type=batterySensor -output battery_generated.go -linecomment
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
	battType       batterySensor = iota // Battery Type
	battPercentage                      // Battery Level
	battTemp                            // Battery Temperature
	battVoltage                         // Battery Voltage
	battEnergy                          // Battery Energy
	battEnergyRate                      // Battery Power
	battState                           // Battery State
	battNativePath                      // Battery Path
	battLevel                           // Battery Level
	battModel                           // Battery Model

	upowerDBusDest         = "org.freedesktop.UPower"
	upowerDBusDeviceDest   = upowerDBusDest + ".Device"
	upowerDBusPath         = "/org/freedesktop/UPower"
	upowerGetDevicesMethod = "org.freedesktop.UPower.EnumerateDevices"

	deviceAddedSignal   = "DeviceAdded"
	deviceRemovedSignal = "DeviceRemoved"

	batteryIcon = "mdi:battery"

	workerID = "battery_sensors"
)

type batterySensor int

var ErrInvalidBattery = errors.New("invalid battery")

// dBusSensorToProps is a map of battery sensors to their D-Bus properties.
var dBusSensorToProps = map[batterySensor]string{
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
var dBusPropToSensor = map[string]batterySensor{
	"Energy":       battEnergy,
	"EnergyRate":   battEnergyRate,
	"Voltage":      battVoltage,
	"Percentage":   battPercentage,
	"Temperatute":  battTemp,
	"State":        battState,
	"BatteryLevel": battLevel,
}

type upowerBattery struct {
	logger   *slog.Logger
	bus      *dbusx.Bus
	id       string
	model    string
	dBusPath dbus.ObjectPath
	sensors  []batterySensor
	battType batteryType
}

// getProp retrieves the property from D-Bus that matches the given battery sensor type.
func (b *upowerBattery) getProp(t batterySensor) (dbus.Variant, error) {
	value, err := dbusx.NewProperty[dbus.Variant](b.bus, string(b.dBusPath), upowerDBusDest, dBusSensorToProps[t]).Get()
	if err != nil {
		return dbus.Variant{}, fmt.Errorf("could not retrieve battery property %s: %w", t.String(), err)
	}

	return value, nil
}

// getSensors retrieves the sensors passed in for a given battery.
func (b *upowerBattery) getSensors(sensors ...batterySensor) chan sensor.Details {
	sensorCh := make(chan sensor.Details, len(sensors))
	defer close(sensorCh)

	for _, batterySensor := range sensors {
		value, err := b.getProp(batterySensor)
		if err != nil {
			b.logger.Warn("Could not retrieve battery sensor.",
				slog.String("sensor", batterySensor.String()),
				slog.Any("error", err))

			continue
		}
		sensorCh <- newBatterySensor(b, batterySensor, value)
	}

	return sensorCh
}

// newBattery creates a battery object that will have a number of properties to
// be treated as sensors in Home Assistant.
func newBattery(bus *dbusx.Bus, logger *slog.Logger, path dbus.ObjectPath) (*upowerBattery, error) {
	battery := &upowerBattery{
		dBusPath: path,
		bus:      bus,
	}

	var (
		variant dbus.Variant
		err     error
	)

	// Get the battery type. Depending on the value, additional sensors will be added.
	variant, err = battery.getProp(battType)
	if err != nil {
		return nil, fmt.Errorf("could not determine battery type: %w", err)
	}
	// Store the battery type.
	battery.battType, err = dbusx.VariantToValue[batteryType](variant)
	if err != nil {
		return nil, fmt.Errorf("could not determine battery type: %w", err)
	}

	// Use the native path D-Bus property for the battery id.
	variant, err = battery.getProp(battNativePath)
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
	variant, err = battery.getProp(battModel)
	if err != nil {
		battery.logger.Warn("Could not determine battery model.")
	}
	// Store the battery model.
	battery.model, err = dbusx.VariantToValue[string](variant)
	if err != nil {
		battery.logger.Warn("Could not determine battery model.")
	}

	// At a minimum, monitor the battery type and the charging state.
	battery.sensors = append(battery.sensors, battState)

	if battery.battType == batteryTypeBattery {
		// Battery has charge percentage, temp and charging rate sensors
		battery.sensors = append(battery.sensors, battPercentage, battTemp, battEnergyRate)
	} else {
		// Battery has a textual level sensor
		battery.sensors = append(battery.sensors, battLevel)
	}

	return battery, nil
}

type upowerBatterySensor struct {
	attributes any
	logger     *slog.Logger
	batteryID  string
	model      string
	linux.Sensor
	sensorType batterySensor
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

//nolint:exhaustive
func (s *upowerBatterySensor) Icon() string {
	switch s.sensorType {
	case battPercentage:
		return batteryPercentIcon(s.Value)
	case battEnergyRate:
		return batteryChargeIcon(s.Value)
	default:
		return batteryIcon
	}
}

//nolint:exhaustive
func (s *upowerBatterySensor) DeviceClass() types.DeviceClass {
	switch s.sensorType {
	case battPercentage:
		return types.DeviceClassBattery
	case battTemp:
		return types.DeviceClassTemperature
	case battEnergyRate:
		return types.DeviceClassPower
	default:
		return 0
	}
}

//nolint:exhaustive
func (s *upowerBatterySensor) StateClass() types.StateClass {
	switch s.sensorType {
	case battPercentage, battTemp, battEnergyRate:
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

	switch s.sensorType {
	case battVoltage, battTemp, battEnergy, battEnergyRate, battPercentage:
		if value, ok := s.Value.(float64); !ok {
			return sensor.StateUnknown
		} else {
			return value
		}
	case battState:
		if value, ok := s.Value.(uint32); !ok {
			return sensor.StateUnknown
		} else {
			return battChargeState(value).String()
		}
	case battLevel:
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

func (s *upowerBatterySensor) Attributes() map[string]any {
	attributes := make(map[string]any)

	attributes["extra_attributes"] = s.attributes

	return attributes
}

//nolint:exhaustive
func (s *upowerBatterySensor) generateAttributes(battery *upowerBattery) {
	switch s.sensorType {
	case battEnergyRate:
		var (
			variant         dbus.Variant
			err             error
			voltage, energy float64
		)

		variant, err = battery.getProp(battVoltage)
		if err != nil {
			s.logger.Warn("Could not retrieve battery voltage.", slog.Any("error", err))
		}

		voltage, err = dbusx.VariantToValue[float64](variant)
		if err != nil {
			s.logger.Warn("Could not retrieve battery voltage.", slog.Any("error", err))
		}

		variant, err = battery.getProp(battEnergy)
		if err != nil {
			s.logger.Warn("Could not retrieve battery energy.", slog.Any("error", err))
		}

		energy, err = dbusx.VariantToValue[float64](variant)
		if err != nil {
			s.logger.Warn("Could not retrieve battery energy.", slog.Any("error", err))
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
	case battPercentage, battLevel:
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
func newBatterySensor(battery *upowerBattery, sensorType batterySensor, value dbus.Variant) *upowerBatterySensor {
	batterySensor := &upowerBatterySensor{
		batteryID:  battery.id,
		model:      battery.model,
		logger:     battery.logger,
		sensorType: sensorType,
	}
	batterySensor.Value = value.Value()
	batterySensor.IsDiagnostic = true
	batterySensor.generateAttributes(battery)

	return batterySensor
}

// monitorBattery will monitor a battery device for any property changes and
// send these as sensors.
func monitorBattery(ctx context.Context, battery *upowerBattery) <-chan sensor.Details {
	sensorCh := make(chan sensor.Details)
	// Create a DBus signal match to watch for property changes for this
	// battery.
	events, err := dbusx.NewWatch(
		dbusx.MatchPath(string(battery.dBusPath)),
		dbusx.MatchPropChanged(),
	).Start(ctx, battery.bus)
	if err != nil {
		battery.logger.Debug("Failed to create D-Bus watch for battery property changes.", slog.Any("error", err))
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
					battery.logger.Warn("Received a battery property change event that could not be understood.", slog.Any("error", err))

					continue
				}

				for prop, value := range props.Changed {
					if s, ok := dBusPropToSensor[prop]; ok {
						sensorCh <- newBatterySensor(battery, s, value)
					}
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
	logger      *slog.Logger
	bus         *dbusx.Bus
	batteryList map[dbus.ObjectPath]context.CancelFunc
	mu          sync.Mutex
}

// ?: implement initial battery sensor retrieval.
func (w *batterySensorWorker) Sensors(_ context.Context) ([]sensor.Details, error) {
	return nil, linux.ErrUnimplemented
}

func (w *batterySensorWorker) Events(ctx context.Context) (chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)

	var wg sync.WaitGroup

	// Get a list of all current connected batteries and monitor them.
	batteries, err := w.getBatteries()
	if err != nil {
		w.logger.Warn("Could not retrieve any battery details from D-Bus.", slog.Any("error", err))
	}

	for _, path := range batteries {
		wg.Add(1)

		go func(path dbus.ObjectPath) {
			defer wg.Done()

			for batterySensor := range w.track(ctx, path) {
				sensorCh <- batterySensor
			}
		}(path)
	}

	wg.Add(1)

	go func() {
		defer wg.Done()

		for batterySensor := range w.monitorBatteryChanges(ctx) {
			sensorCh <- batterySensor
		}
	}()

	go func() {
		defer close(sensorCh)
		wg.Wait()
	}()

	return sensorCh, nil
}

// getBatteries is a helper function to retrieve all of the known batteries
// connected to the system.
func (w *batterySensorWorker) getBatteries() ([]dbus.ObjectPath, error) {
	batteryList, err := dbusx.GetData[[]dbus.ObjectPath](w.bus, upowerDBusPath, upowerDBusDest, upowerGetDevicesMethod)
	if err != nil {
		return nil, err
	}

	return batteryList, nil
}

func (w *batterySensorWorker) track(ctx context.Context, batteryPath dbus.ObjectPath) <-chan sensor.Details {
	sensorCh := make(chan sensor.Details)

	var wg sync.WaitGroup

	battery, err := newBattery(w.bus, w.logger, batteryPath)
	if err != nil {
		w.logger.Warn("Cannot monitor battery.",
			slog.Any("path", batteryPath),
			slog.Any("error", err))

		return sensorCh
	}

	battCtx, cancelFunc := context.WithCancel(ctx)

	w.mu.Lock()
	w.batteryList[batteryPath] = cancelFunc
	w.mu.Unlock()

	wg.Add(1)

	go func() {
		defer wg.Done()

		for prop := range battery.getSensors(battery.sensors...) {
			sensorCh <- prop
		}
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()

		for battery := range monitorBattery(battCtx, battery) {
			sensorCh <- battery
		}
	}()

	go func() {
		defer close(sensorCh)
		wg.Wait()
	}()

	return sensorCh
}

func (w *batterySensorWorker) remove(batteryPath dbus.ObjectPath) {
	if cancelFunc, ok := w.batteryList[batteryPath]; ok {
		cancelFunc()
		w.mu.Lock()
		delete(w.batteryList, batteryPath)
		w.mu.Unlock()
	}
}

// monitorBatteryChanges monitors for battery devices being added/removed from
// the system and will start/stop monitory each battery as appropriate.
func (w *batterySensorWorker) monitorBatteryChanges(ctx context.Context) <-chan sensor.Details {
	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(upowerDBusPath),
		dbusx.MatchInterface(upowerDBusDest),
		dbusx.MatchMembers(deviceAddedSignal, deviceRemovedSignal),
	).Start(ctx, w.bus)
	if err != nil {
		w.logger.Debug("Unable to set-up D-Bus watch for battery changes.", slog.Any("error", err))

		return nil
	}

	sensorCh := make(chan sensor.Details)

	// events, err := dbusx.NewWatch(
	// 	dbusx.MatchPath(upowerDBusPath),
	// 	dbusx.MatchInterface(upowerDBusDest),
	// 	dbusx.MatchMember(deviceAddedSignal, deviceRemovedSignal),
	// ).Start(ctx, w.bus)
	// if err != nil {
	// 	w.logger.Debug("Failed to create D-Bus watch for battery additions/removals.", "error", err.Error())
	// 	close(sensorCh)

	// 	return sensorCh
	// }

	go func() {
		w.logger.Debug("Monitoring for battery additions/removals.")

		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				w.logger.Debug("Stopped monitoring for batteries.")

				return
			case event := <-triggerCh:
				batteryPath, validBatteryPath := event.Content[0].(dbus.ObjectPath)
				if !validBatteryPath {
					continue
				}

				switch {
				case strings.Contains(event.Signal, deviceAddedSignal):
					go func() {
						for s := range w.track(ctx, batteryPath) {
							sensorCh <- s
						}
					}()
				case strings.Contains(event.Signal, deviceRemovedSignal):
					w.remove(batteryPath)
				}
			}
		}
	}()

	return sensorCh
}

func NewBatteryWorker(ctx context.Context) (*linux.SensorWorker, error) {
	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return nil, linux.ErrNoSystemBus
	}

	return &linux.SensorWorker{
			Value: &batterySensorWorker{
				logger:      logging.FromContext(ctx).With(slog.String("worker", workerID)),
				bus:         bus,
				batteryList: make(map[dbus.ObjectPath]context.CancelFunc),
			},
			WorkerID: workerID,
		},
		nil
}
