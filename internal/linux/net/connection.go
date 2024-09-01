// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package net

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"github.com/godbus/dbus/v5"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	connectionIDProp         = dbusNMActiveConnIntr + ".Id"
	connectionTypeProp       = dbusNMActiveConnIntr + ".Type"
	connectionDevicesProp    = dbusNMActiveConnIntr + ".Devices"
	connectionStateProp      = dbusNMActiveConnIntr + "." + statePropName
	connectionIPv4ConfigProp = dbusNMActiveConnIntr + "." + ipv4ConfigPropName
	connectionIPv6ConfigProp = dbusNMActiveConnIntr + "." + ipv6ConfigPropName

	wirelessDeviceInterface = dBusNMObj + ".Device.Wireless"

	activeAPPropName = "ActiveAccessPoint"
)

var (
	ErrUnknownProp      = errors.New("unknown or invalid property")
	ErrUnsupportedValue = errors.New("unsupported state value")
)

type connectionSensor interface {
	sensor.Details
	setState(value any) error
	updateState() error
}

type connection struct {
	devicesProp *dbusx.Property[[]dbus.ObjectPath]
	logger      *slog.Logger
	doneCh      chan struct{}
	name        string
	connType    string
	path        dbus.ObjectPath
}

// newConnection sets up an object that tracks connection state. It both static
// (stored in the object) and dynamic (fetched from D-Bus as needed) properties
// of a connection.
func newConnection(bus *dbusx.Bus, path dbus.ObjectPath) (*connection, error) {
	// Get the connection name.
	name, err := dbusx.NewProperty[string](bus, string(path), dBusNMObj, connectionIDProp).Get()
	if err != nil {
		return nil, fmt.Errorf("could not determine connection name: %w", err)
	}

	connType, err := dbusx.NewProperty[string](bus, string(path), dBusNMObj, connectionTypeProp).Get()
	if err != nil {
		return nil, fmt.Errorf("could not determine connection type: %w", err)
	}

	conn := &connection{
		path:        path,
		name:        name,
		connType:    connType,
		devicesProp: dbusx.NewProperty[[]dbus.ObjectPath](bus, string(path), dBusNMObj, connectionDevicesProp),
		logger:      slog.With(slog.String("connection", name)),
		doneCh:      make(chan struct{}),
	}

	return conn, nil
}

// monitor will set up a D-Bus watch on the connection path for
// connection property changes and send those back through the returned channel
// as sensors.
func (c *connection) monitor(ctx context.Context, bus *dbusx.Bus) <-chan sensor.Details {
	sensorCh := make(chan sensor.Details)

	// Monitor connection properties.
	go func() {
		for sensor := range c.monitorConnection(ctx, bus) {
			sensorCh <- sensor
		}
	}()

	// If the connection is a wifi connection, monitor wifi properties.
	if c.connType == "802-11-wireless" {
		go func() {
			for sensor := range c.monitorWifi(ctx, bus) {
				sensorCh <- sensor
			}
		}()
	}

	return sensorCh
}

// monitorConnection sets up the D-Bus watch for connection property changes.
//
//nolint:gocognit,gocyclo,cyclop
//revive:disable:function-length
func (c *connection) monitorConnection(ctx context.Context, bus *dbusx.Bus) <-chan sensor.Details {
	sensorCh := make(chan sensor.Details)

	// Create sensors for monitored properties.
	stateSensor := newConnectionStateSensor(bus, string(c.path), c.name)
	ipv4Sensor := newConnectionAddrSensor(bus, ipv4, string(c.path), c.name)
	ipv6Sensor := newConnectionAddrSensor(bus, ipv6, string(c.path), c.name)
	// Update their states.
	for _, connSensor := range []connectionSensor{stateSensor, ipv4Sensor, ipv6Sensor} {
		if err := connSensor.updateState(); err != nil {
			c.logger.Debug("Could not update sensor.", slog.String("sensor", connSensor.Name()), slog.Any("error", err))
		}
	}
	// Send initial states as sensors
	go func() {
		for _, connSensor := range []connectionSensor{stateSensor, ipv4Sensor, ipv6Sensor} {
			sensorCh <- connSensor
		}
	}()

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(string(c.path)),
		dbusx.MatchPropChanged(),
	).Start(ctx, bus)
	if err != nil {
		c.logger.Debug("Could not start D-Bus connection property watch.", slog.Any("error", err))
		close(sensorCh)

		return sensorCh
	}

	go func() {
		defer close(sensorCh)

		c.logger.Debug("Started monitoring connection properties.")

		for {
			select {
			case <-c.doneCh: // Connection offline.
				c.logger.Debug("Stopped monitoring connection properties (connection offline).")

				return
			case <-ctx.Done(): // Agent shutting down.
				c.logger.Debug("Stopped monitoring connection properties (agent shutdown).")

				return
			case event := <-triggerCh: // Connection property changed.
				props, err := dbusx.ParsePropertiesChanged(event.Content)
				if err != nil {
					continue
				}
				// Ignore device statistics.
				if props.Interface == "org.freedesktop.NetworkManager.Device.Statistics" {
					continue
				}

				for prop, value := range props.Changed {
					switch {
					case prop == statePropName && props.Interface == dbusNMActiveConnIntr: // State changed.
						if err := stateSensor.setState(value); err != nil {
							c.logger.Warn("Could not update connection state sensor.", slog.Any("error", err))
						} else {
							// Send the connection state as a sensor.
							sensorCh <- stateSensor
						}

						if state, ok := stateSensor.State().(connState); ok && state == connOffline {
							close(c.doneCh)
						}
					case prop == ipv4ConfigPropName:
						if err := ipv4Sensor.setState(value); err != nil {
							c.logger.Warn("Could not parse updated ipv4 address.", slog.Any("error", err))
						} else {
							sensorCh <- ipv4Sensor
						}
					case prop == ipv6ConfigPropName: // IP addresses changed.
						if err := ipv6Sensor.setState(value); err != nil {
							c.logger.Warn("Could not parse updated ipv6 address.", slog.Any("error", err))
						} else {
							sensorCh <- ipv6Sensor
						}
					default:
						c.logger.Debug("Unhandled property changed.",
							slog.String("interface", props.Interface),
							slog.String("property", prop),
							slog.Any("value", value.Value()))
					}
				}
			}
		}
	}()

	return sensorCh
}

//nolint:cyclop
func (c *connection) monitorWifi(ctx context.Context, bus *dbusx.Bus) <-chan sensor.Details {
	triggerCh := make(chan dbusx.Trigger)
	sensorCh := make(chan sensor.Details)

	// Get and send initial values for wifi props.
	go func() {
		for _, ap := range c.getWifiAPs(bus) {
			for _, wifiSensor := range getWifiSensors(bus, string(ap)) {
				sensorCh <- wifiSensor
			}
		}
	}()

	go func() {
		c.watchWifiDevice(ctx, bus, triggerCh)
	}()

	var apCancelFunc context.CancelFunc

	apCancelFunc = c.watchAPs(ctx, bus, triggerCh)

	go func() {
		defer close(sensorCh)

		c.logger.Debug("Started monitoring wifi properties.")

		for {
			select {
			case <-c.doneCh: // Connection offline.
				c.logger.Debug("Stopped monitoring wifi properties (connection offline).")

				return
			case <-ctx.Done(): // Agent shutting down.
				c.logger.Debug("Stopped monitoring wifi properties (agent shutdown).")

				return
			case event := <-triggerCh: // Wifi property changed.
				props, err := dbusx.ParsePropertiesChanged(event.Content)
				if err != nil {
					continue
				}
				// Ignore device statistics.
				if props.Interface == "org.freedesktop.NetworkManager.Device.Statistics" {
					continue
				}

				for prop, value := range props.Changed {
					switch {
					case prop == activeAPPropName: // Access point changed.
						apCancelFunc()
						apCancelFunc = c.watchAPs(ctx, bus, triggerCh)
					case slices.Contains(apPropList, prop): // Wifi property changed.
						sensorCh <- newWifiSensor(prop, value.Value())
					}
				}
			}
		}
	}()

	return sensorCh
}

// watchAPs will get the current active access points for a wireless connections
// and set up a D-Bus watch on the APs property changes. It returns a cancelFunc
// that can be used to cancel the watch. Changed properties are sent back
// through the passed dbusx.Trigger channel.
func (c *connection) watchAPs(ctx context.Context, bus *dbusx.Bus, triggerCh chan dbusx.Trigger) context.CancelFunc {
	activeAccessPoints := c.getWifiAPs(bus)
	apCtx, apCancel := context.WithCancel(ctx)

	for _, ap := range activeAccessPoints {
		c.watchAP(apCtx, bus, string(ap), triggerCh)
	}

	return apCancel
}

// watchWifiDevice will create a watch on the device properties for a wireless
// connection. Changed properties are sent back through the passed dbusx.Trigger
// channel.
func (c *connection) watchWifiDevice(ctx context.Context, bus *dbusx.Bus, outCh chan dbusx.Trigger) {
	devices, err := c.devicesProp.Get()
	if err != nil {
		c.logger.Debug("could not retrieve wireless devices for connection from D-Bus", slog.Any("error", err))

		return
	}

	triggerChs := make([]<-chan dbusx.Trigger, 0, len(devices))

	for _, devicePath := range devices {
		// Monitor device properties.
		devicePropCh, err := dbusx.NewWatch(
			dbusx.MatchPath(string(devicePath)),
			dbusx.MatchPropChanged(),
		).Start(ctx, bus)
		if err != nil {
			c.logger.Debug("Could not start D-Bus wifi property watch", slog.Any("error", err))

			return
		}

		triggerChs = append(triggerChs, devicePropCh)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-mergeCh(ctx, triggerChs...):
			outCh <- event
		}
	}
}

// getWifiAPs returns a slice of dbus.ObjectPath representing all the active
// access points the connection is using.
func (c *connection) getWifiAPs(bus *dbusx.Bus) []dbus.ObjectPath {
	devices, err := c.devicesProp.Get()
	if err != nil {
		c.logger.Debug("could not retrieve wireless devices for connection from D-Bus", slog.Any("error", err))

		return nil
	}

	aps := make([]dbus.ObjectPath, 0, len(devices))

	for _, devicePath := range devices {
		apPath, err := dbusx.NewProperty[dbus.ObjectPath](bus, string(devicePath), dBusNMObj, wirelessDeviceInterface+"."+activeAPPropName).Get()
		if err != nil {
			c.logger.Debug("no ap path", slog.Any("error", err))

			continue
		}

		aps = append(aps, apPath)
	}

	return aps
}

// watchAP will set up a D-Bus watch for a connection on its active wireless
// access point.
func (c *connection) watchAP(ctx context.Context, bus *dbusx.Bus, apPath string, outCh chan dbusx.Trigger) {
	apPropCh, err := dbusx.NewWatch(
		dbusx.MatchPath(apPath),
		dbusx.MatchPropChanged(),
	).Start(ctx, bus)
	if err != nil {
		c.logger.Debug("Could not start D-Bus access point property watch.", slog.Any("error", err))

		return
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-apPropCh:
				outCh <- event
			}
		}
	}()
}

func mergeCh[T any](ctx context.Context, inCh ...<-chan T) chan T {
	var wg sync.WaitGroup

	outCh := make(chan T)

	// Start an output goroutine for each input channel in sensorCh.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(ch <-chan T) { //nolint:varnamelen
		defer wg.Done()

		if ch == nil {
			return
		}

		for n := range ch {
			select {
			case outCh <- n:
			case <-ctx.Done():
				return
			}
		}
	}

	wg.Add(len(inCh))

	for _, c := range inCh {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(outCh)
	}()

	return outCh
}
