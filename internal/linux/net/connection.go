// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package net

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/godbus/dbus/v5"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/models"
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
func (c *connection) monitor(ctx context.Context, bus *dbusx.Bus) <-chan models.Entity {
	sensorCh := make(chan models.Entity)

	// Monitor connection properties.
	go func() {
		c.logger.Debug("Monitoring connection.")

		for sensor := range c.monitorConnection(ctx, bus) {
			sensorCh <- sensor
		}

		c.logger.Debug("Unmonitoring connection.")
	}()

	// If the connection is a wifi connection, monitor wifi properties.
	if c.connType == "802-11-wireless" {
		go func() {
			c.logger.Debug("Monitoring WiFi.")

			for sensor := range c.monitorWifi(ctx, bus) {
				sensorCh <- sensor
			}

			c.logger.Debug("Unmonitoring WiFi.")
		}()
	}

	return sensorCh
}

// monitorConnection sets up the D-Bus watch for connection property changes.
//
//nolint:gocognit,funlen
func (c *connection) monitorConnection(ctx context.Context, bus *dbusx.Bus) <-chan models.Entity {
	var (
		stateSensor *connectionStateSensor
		entity      *models.Entity
		err         error
	)

	sensorCh := make(chan models.Entity)
	monitorCtx, monitorCancel := context.WithCancel(ctx)

	// Create sensors for monitored properties.
	stateSensor, err = newConnectionStateSensor(bus, string(c.path), c.name)
	if err != nil {
		c.logger.Debug("Could not update sensor.",
			slog.String("sensor", stateSensor.name),
			slog.Any("error", err))
	}

	// Send initial states as sensors
	go func() {
		if entity, err = stateSensor.generateEntity(ctx); err != nil {
			c.logger.Debug("Could not generate sensor from connection state.",
				slog.String("sensor", stateSensor.name),
				slog.Any("error", err))
		} else {
			sensorCh <- *entity
		}
	}()

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(string(c.path)),
		dbusx.MatchPropChanged(),
	).Start(monitorCtx, bus)
	if err != nil {
		c.logger.Debug("Could not start D-Bus connection property watch.", slog.Any("error", err))
		monitorCancel()
		close(sensorCh)

		return sensorCh
	}

	go func() {
		defer close(sensorCh)
		defer monitorCancel()
		defer close(c.doneCh)
		c.logger.Debug("Stated watching for connection property updates.")

		for {
			select {
			case <-ctx.Done():
				c.logger.Debug("Stopped watching for connection property updates.")

				return
			case event := <-triggerCh:
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
							if entity, err := stateSensor.generateEntity(ctx); err != nil {
								c.logger.Debug("Could not generate sensor from connection state.",
									slog.String("sensor", stateSensor.name),
									slog.Any("error", err))
							} else {
								sensorCh <- *entity
							}
						}
					default:
						c.logger.Debug("Unhandled property changed.",
							slog.String("interface", props.Interface),
							slog.String("property", prop),
							slog.Any("value", value.Value()))
					}
				}
			}

			if stateSensor.state == connOffline.String() {
				break
			}
		}
	}()

	return sensorCh
}

// monitorWifi will monitor wifi connection properties.
//
//nolint:gocognit
func (c *connection) monitorWifi(ctx context.Context, bus *dbusx.Bus) <-chan models.Entity {
	triggerCh := make(chan dbusx.Trigger)
	sensorCh := make(chan models.Entity)
	monitorCtx, monitorCancel := context.WithCancel(ctx)

	// Get and send initial values for wifi props.
	go func() {
		for _, ap := range c.getWifiAPs(bus) {
			for _, wifiSensor := range getWifiSensors(ctx, bus, string(ap)) {
				sensorCh <- wifiSensor
			}
		}
	}()

	go func() {
		c.watchAccessPointProps(monitorCtx, bus, triggerCh)
	}()

	go func() {
		defer close(sensorCh)
		defer monitorCancel()

		c.logger.Debug("Started monitoring wifi properties.")

		for {
			select {
			case <-c.doneCh: // Connection offline.
				return
			case <-ctx.Done(): // Agent shutting down.
				return
			case event := <-triggerCh: // Wifi property changed.
				props, err := dbusx.ParsePropertiesChanged(event.Content)
				if err != nil {
					continue
				}

				for prop, value := range props.Changed {
					if slices.Contains(apPropList, prop) { // Wifi property changed.
						entity, err := newWifiSensor(ctx, prop, value.Value())
						if err != nil {
							logging.FromContext(ctx).Warn("Could not generate new wifi property sensor.",
								slog.Any("error", err))
							continue
						}

						sensorCh <- *entity
					}
				}
			}
		}
	}()

	return sensorCh
}

// watchAccessPointProps sets up the watches for changes to access points and
// their properties.
func (c *connection) watchAccessPointProps(ctx context.Context, bus *dbusx.Bus, triggerCh chan dbusx.Trigger) {
	apCh := make(chan dbus.ObjectPath)
	defer close(apCh)

	apWatchCtx, apWatchCancel := context.WithCancel(ctx)
	defer apWatchCancel()

	devices, err := c.devicesProp.Get()
	if err != nil {
		c.logger.Debug("Could not retrieve wireless devices for connection from D-Bus", slog.Any("error", err))

		return
	}

	go func() {
		// Monitor access point changes on devices.
		for _, devicePath := range devices {
			c.monitorDeviceAccessPoint(apWatchCtx, bus, string(devicePath), apCh)
		}
		// Send the current active access points.
		for _, ap := range c.getWifiAPs(bus) {
			apCh <- ap
		}
	}()

	var apPropCancel context.CancelFunc

	for {
		select {
		case <-c.doneCh:
			return
		case <-ctx.Done():
			return
		case accessPoint := <-apCh:
			// If there was a previous ap watch, cancel it.
			if apPropCancel != nil {
				apPropCancel()
			}
			// Watch this ap for prop changes.
			apPropCancel = c.watchAP(ctx, bus, string(accessPoint), triggerCh)
		}
	}
}

// monitorDeviceAccessPoint starts a D-Bus watch for changes to the active
// access point for a device.
func (c *connection) monitorDeviceAccessPoint(ctx context.Context, bus *dbusx.Bus, devicePath string, outCh chan dbus.ObjectPath) {
	monitorCtx, monitorCancel := context.WithCancel(ctx)

	// Monitor the active access point property.
	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(devicePath),
		dbusx.MatchInterface(wirelessDeviceInterface),
		dbusx.MatchMembers(activeAPPropName),
	).Start(monitorCtx, bus)
	if err != nil {
		c.logger.Debug("Could not monitor device access point.", slog.Any("error", err))
		monitorCancel()

		return
	}

	go func() {
		defer monitorCancel()

		for {
			select {
			case <-c.doneCh:
				return
			case <-ctx.Done():
				return
			case event := <-triggerCh:
				values, err := dbusx.ParseValueChange[dbus.ObjectPath](event.Content)
				if err != nil {
					c.logger.Debug("Could not parse changed access point prop.", slog.Any("error", err))

					continue
				}
				outCh <- values.New
			}
		}
	}()
}

// getWifiAPs returns a slice of dbus.ObjectPath representing all the active
// access points the connection is using.
func (c *connection) getWifiAPs(bus *dbusx.Bus) []dbus.ObjectPath {
	devices, err := c.devicesProp.Get()
	if err != nil {
		c.logger.Debug("Could not retrieve active access points.", slog.Any("error", err))

		return nil
	}

	aps := make([]dbus.ObjectPath, 0, len(devices))

	for _, devicePath := range devices {
		apPath, err := dbusx.NewProperty[dbus.ObjectPath](bus, string(devicePath), dBusNMObj, wirelessDeviceInterface+"."+activeAPPropName).Get()
		if err != nil {
			continue
		}

		aps = append(aps, apPath)
	}

	return aps
}

// watchAP will set up a D-Bus watch for a connection on its active wireless
// access point and send any access point property changes to the given trigger
// channel. It returns a context.CancelFunc that can be used to stop the watch.
func (c *connection) watchAP(ctx context.Context, bus *dbusx.Bus, apPath string, outCh chan dbusx.Trigger) context.CancelFunc {
	watchCtx, watchCancel := context.WithCancel(ctx)

	apPropCh, err := dbusx.NewWatch(
		dbusx.MatchPath(apPath),
		dbusx.MatchPropChanged(),
	).Start(watchCtx, bus)
	if err != nil {
		c.logger.Debug("Could not start D-Bus access point property watch.", slog.Any("error", err))

		return watchCancel
	}

	go func() {
		defer watchCancel()

		for {
			select {
			case <-c.doneCh:
				return
			case event := <-apPropCh:
				outCh <- event
			}
		}
	}()

	return watchCancel
}
