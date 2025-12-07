// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//go:generate go tool golang.org/x/tools/cmd/stringer -type=connState,connIcon -output nm_connection.gen.go -linecomment
package network

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/iancoleman/strcase"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/logging"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/class"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

const (
	dBusNMPath           = "/org/freedesktop/NetworkManager"
	dBusNMObj            = "org.freedesktop.NetworkManager"
	dbusNMActiveConnPath = dBusNMPath + "/ActiveConnection"
	dbusNMActiveConnIntr = dBusNMObj + ".Connection.Active"

	ipv4ConfigPropName = "Ip4Config"
	ipv6ConfigPropName = "Ip6Config"
	statePropName      = "State"

	netConnWorkerID   = "network_connection_sensors"
	netConnWorkerDesc = "NetworkManager connection status"
	netConnPrefID     = prefPrefix + "connections"
)

var _ workers.EntityWorker = (*ConnectionsWorker)(nil)

type ConnectionsWorker struct {
	*models.WorkerMetadata

	bus   *dbusx.Bus
	list  map[string]*connection
	prefs *CommonPreferences
	mu    sync.Mutex
}

// NewNMConnectionWorker creates a new sensor worker that monitors NetworkManager through D-Bus for new connections.
func NewNMConnectionWorker(ctx context.Context) (workers.EntityWorker, error) {
	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return nil, fmt.Errorf("get system bus: %w", linux.ErrNoSystemBus)
	}

	worker := &ConnectionsWorker{
		WorkerMetadata: models.SetWorkerMetadata(netConnWorkerID, netConnWorkerDesc),
		bus:            bus,
		list:           make(map[string]*connection),
	}

	defaultPrefs := &CommonPreferences{
		IgnoredDevices: defaultIgnoredDevices,
	}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(netConnPrefID, defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	return worker, nil
}

func (w *ConnectionsWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	sensorCh := make(chan models.Entity)
	connCtx, connCancel := context.WithCancel(ctx)

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPathNamespace(dbusNMActiveConnPath),
		dbusx.MatchInterface(dbusNMActiveConnIntr),
		dbusx.MatchMembers("StateChanged"),
	).Start(connCtx, w.bus)
	if err != nil {
		close(sensorCh)
		connCancel()

		return sensorCh, fmt.Errorf("watch connection state: %w", err)
	}

	go func() {
		defer close(sensorCh)

		for event := range triggerCh {
			connectionPath := dbus.ObjectPath(event.Path)
			// If this connection is in the process of deactivating, don't
			// start tracking it.
			if state, stateChange := event.Content[0].(uint32); stateChange {
				if state > uint32(connOnline) {
					continue
				}
			}
			// Track all activating/new connections.
			if err = w.handleConnection(connCtx, connectionPath, sensorCh); err != nil {
				slogctx.FromCtx(ctx).Debug("Could not monitor connection.",
					slog.String("dbus_path", string(connectionPath)),
					slog.Any("error", err))
			}
		}
	}()

	go func() {
		defer connCancel()
		<-ctx.Done()
	}()

	// monitor all current active connections
	connectionlist, err := dbusx.NewProperty[[]dbus.ObjectPath](
		w.bus,
		dBusNMPath,
		dBusNMObj,
		dBusNMObj+".ActiveConnections",
	).Get()
	if err != nil {
		return nil, fmt.Errorf("list active connections: %w", err)
	}
	for _, path := range connectionlist {
		if err := w.handleConnection(connCtx, path, sensorCh); err != nil {
			slogctx.FromCtx(ctx).Debug("Could not monitor connection.", slog.Any("error", err))
		}
	}

	return sensorCh, nil
}

func (w *ConnectionsWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *ConnectionsWorker) isTracked(id string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, found := w.list[id]; found {
		return true
	}

	return false
}

func (w *ConnectionsWorker) handleConnection(
	ctx context.Context,
	path dbus.ObjectPath,
	sensorCh chan models.Entity,
) error {
	conn, err := newConnection(ctx, w.bus, path)
	if err != nil {
		return fmt.Errorf("could not create connection: %w", err)
	}
	// Ignore loopback or already tracked connections.
	if conn.name == "lo" || w.isTracked(conn.name) {
		return nil
	}
	// Ignore user-defined devices.
	if slices.ContainsFunc(w.prefs.IgnoredDevices, func(e string) bool {
		return strings.HasPrefix(conn.name, e)
	}) {
		return nil
	}

	w.mu.Lock()
	w.list[conn.name] = conn
	w.mu.Unlock()

	// Start monitoring the connection. Pass any sensor updates from the
	// connection through the sensor channel.
	go func() {
		var wg sync.WaitGroup
		wg.Go(func() {
			// Monitor wired connection properties.
			for sensor := range conn.monitorWired(ctx, w.bus) {
				sensorCh <- sensor
			}
		})
		if conn.connType == "802-11-wireless" {
			wg.Go(func() {
				// Monitor wireless connection properties.
				for sensor := range conn.monitorWireless(ctx, w.bus) {
					sensorCh <- sensor
				}
			})
		}

		// Monitor wired connection properties.
		slogctx.FromCtx(ctx).Debug("Started monitoring connection.",
			slog.String("connection", conn.name),
		)
		wg.Wait()
		w.mu.Lock()
		delete(w.list, conn.name)
		w.mu.Unlock()
		slogctx.FromCtx(ctx).Debug("Stopped monitoring connection.",
			slog.String("connection", conn.name),
		)
	}()

	return nil
}

const (
	connUnknown      connState = iota // Unknown
	connActivating                    // Activating
	connOnline                        // Online
	connDeactivating                  // Deactivating
	connOffline                       // Offline
)

// connState represents the connection state.
type connState uint32

const (
	iconUnknown      connIcon = iota // mdi:help-network
	iconActivating                   // mdi:plus-network
	iconOnline                       // mdi:network
	iconDeactivating                 // mdi:network-minus
	iconOffline                      // mdi:network-off
)

// connIcon is an icon representation of the connection state.
type connIcon uint32

// connectionState tracks properties about a connection.
type connectionState struct {
	name      string
	state     string
	icon      string
	stateProp *dbusx.Property[connState]
}

func newConnectionState(bus *dbusx.Bus, connectionPath, connectionName string) (*connectionState, error) {
	conn := &connectionState{
		name:      connectionName,
		stateProp: dbusx.NewProperty[connState](bus, connectionPath, dBusNMObj, connectionStateProp),
	}
	state, err := conn.stateProp.Get()
	if err != nil {
		return nil, fmt.Errorf("get connection state: %w", err)
	}
	conn.state = state.String()
	conn.icon = connIcon(state).String()
	return conn, nil
}

func (c *connectionState) createSensor(ctx context.Context) models.Entity {
	return sensor.NewSensor(ctx,
		sensor.WithName(c.name+" Connection State"),
		sensor.WithID(strcase.ToSnake(c.name)+"_connection_state"),
		sensor.WithDataSourceAttribute(linux.DataSrcDBus),
		sensor.WithState(c.state),
		sensor.WithIcon(c.icon),
	)
}

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
	doneCh      chan struct{}
	name        string
	connType    string
	path        dbus.ObjectPath
}

// newConnection sets up an object that tracks connection state. It both static
// (stored in the object) and dynamic (fetched from D-Bus as needed) properties
// of a connection.
func newConnection(ctx context.Context, bus *dbusx.Bus, path dbus.ObjectPath) (*connection, error) {
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
		doneCh:      make(chan struct{}),
	}

	return conn, nil
}

// monitorWired sets up the D-Bus watch for connection property changes.
//
//nolint:gocognit
func (c *connection) monitorWired(ctx context.Context, bus *dbusx.Bus) <-chan models.Entity {
	var (
		sensor *connectionState
		err    error
	)

	sensorCh := make(chan models.Entity)
	monitorCtx, monitorCancel := context.WithCancel(ctx)

	// Create sensors for monitored properties.
	sensor, err = newConnectionState(bus, string(c.path), c.name)
	if err != nil {
		slogctx.FromCtx(ctx).Debug("Could not create wired connection sensor.",
			slog.String("connection", c.name),
			slog.Any("error", err))
	}

	// Send initial states as sensors
	go func() {
		sensorCh <- sensor.createSensor(ctx)
	}()

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(string(c.path)),
		dbusx.MatchPropChanged(),
	).Start(monitorCtx, bus)
	if err != nil {
		slogctx.FromCtx(ctx).Debug("Could watch wired connection properties.",
			slog.String("connection", c.name),
			slog.Any("error", err))
		monitorCancel()
		close(sensorCh)

		return sensorCh
	}

	go func() {
		defer close(sensorCh)
		defer monitorCancel()
		defer close(c.doneCh)

		for {
			select {
			case <-ctx.Done():
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
						if state, err := dbusx.VariantToValue[connState](value); err == nil {
							sensor.state = state.String()
							sensor.icon = connIcon(state).String()
							sensorCh <- sensor.createSensor(ctx)
						} else if state, err := dbusx.VariantToValue[uint32](value); err == nil {
							sensor.state = connState(state).String()
							sensor.icon = connIcon(state).String()
							sensorCh <- sensor.createSensor(ctx)
						} else {
							slogctx.FromCtx(ctx).Warn("Could not update wired connection state sensor.",
								slog.String("connection", c.name),
								slog.Any("error", ErrUnsupportedValue))
							continue
						}
					default:
						slogctx.FromCtx(ctx).
							Log(ctx, logging.LevelTrace, "Unhandled wired connection property changed.",
								slog.String("connection", c.name),
								slog.String("interface", props.Interface),
								slog.String("property", prop),
								slog.Any("value", value.Value()))
					}
				}
			}

			if sensor.state == connOffline.String() {
				break
			}
		}
	}()

	return sensorCh
}

// monitorWireless will monitor wifi connection properties.
//
//nolint:gocognit
func (c *connection) monitorWireless(ctx context.Context, bus *dbusx.Bus) <-chan models.Entity {
	triggerCh := make(chan dbusx.Trigger)
	sensorCh := make(chan models.Entity)
	monitorCtx, monitorCancel := context.WithCancel(ctx)

	// Get and send initial values for wifi props.
	go func() {
		for _, ap := range c.getWifiAPs(ctx, bus) {
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
						sensorCh <- newWifiSensor(ctx, prop, value.Value())
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
		slogctx.FromCtx(ctx).Debug("Could get wireless connection access point properties",
			slog.String("connection", c.name),
			slog.Any("error", err))

		return
	}

	go func() {
		// Monitor access point changes on devices.
		for _, devicePath := range devices {
			c.monitorDeviceAccessPoint(apWatchCtx, bus, string(devicePath), apCh)
		}
		// Send the current active access points.
		for _, ap := range c.getWifiAPs(ctx, bus) {
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
func (c *connection) monitorDeviceAccessPoint(
	ctx context.Context,
	bus *dbusx.Bus,
	devicePath string,
	outCh chan dbus.ObjectPath,
) {
	monitorCtx, monitorCancel := context.WithCancel(ctx)

	// Monitor the active access point property.
	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(devicePath),
		dbusx.MatchInterface(wirelessDeviceInterface),
		dbusx.MatchMembers(activeAPPropName),
	).Start(monitorCtx, bus)
	if err != nil {
		slogctx.FromCtx(ctx).Debug("Could not wireless connection access point.",
			slog.String("connection", c.name),
			slog.Any("error", err))
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
					slogctx.FromCtx(ctx).Debug("Unhandled access point event.",
						slog.String("connection", c.name),
						slog.Any("error", err),
					)
					continue
				}
				outCh <- values.New
			}
		}
	}()
}

// getWifiAPs returns a slice of dbus.ObjectPath representing all the active
// access points the connection is using.
func (c *connection) getWifiAPs(ctx context.Context, bus *dbusx.Bus) []dbus.ObjectPath {
	devices, err := c.devicesProp.Get()
	if err != nil {
		slogctx.FromCtx(ctx).
			Debug("Could not retrieve wireless connection active access points.",
				slog.String("connection", c.name),
				slog.Any("error", err))

		return nil
	}

	aps := make([]dbus.ObjectPath, 0, len(devices))

	for _, devicePath := range devices {
		apPath, err := dbusx.NewProperty[dbus.ObjectPath](
			bus,
			string(devicePath),
			dBusNMObj,
			wirelessDeviceInterface+"."+activeAPPropName,
		).Get()
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
func (c *connection) watchAP(
	ctx context.Context,
	bus *dbusx.Bus,
	apPath string,
	outCh chan dbusx.Trigger,
) context.CancelFunc {
	watchCtx, watchCancel := context.WithCancel(ctx)

	apPropCh, err := dbusx.NewWatch(
		dbusx.MatchPath(apPath),
		dbusx.MatchPropChanged(),
	).Start(watchCtx, bus)
	if err != nil {
		slogctx.FromCtx(ctx).Debug("Could not watch wireless connection access point properties.",
			slog.String("connection", c.name),
			slog.Any("error", err))

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

const (
	accessPointInterface = dBusNMObj + ".AccessPoint"

	ssidPropName       = "Ssid"
	hwAddrPropName     = "HwAddress"
	maxBitRatePropName = "MaxBitrate"
	freqPropName       = "Frequency"
	strPropName        = "Strength"
	bandwidthPropName  = "Bandwidth"

	unknownState = "Unknown"
)

var apPropList = []string{
	ssidPropName,
	hwAddrPropName,
	maxBitRatePropName,
	freqPropName,
	strPropName,
	bandwidthPropName,
}

var ErrNewWifiPropSensor = errors.New("could not create wifi property sensor")

func newWifiSensor(ctx context.Context, prop string, value any) models.Entity {
	var (
		name, id, units string
		deviceClass     class.SensorDeviceClass
		stateClass      class.SensorStateClass
	)

	icon := "mdi:wifi"

	switch prop {
	case ssidPropName:
		name = "Wi-Fi SSID"
		id = "wi_fi_ssid"
	case hwAddrPropName:
		name = "Wi-Fi BSSID"
		id = "wi_fi_bssid"
	case maxBitRatePropName:
		name = "Wi-Fi Link Speed"
		id = "wi_fi_link_speed"
		units = "kB/s"
		deviceClass = class.SensorClassDataRate
		stateClass = class.StateMeasurement
	case freqPropName:
		name = "Wi-Fi Frequency"
		id = "wi_fi_frequency"
		units = "MHz"
		deviceClass = class.SensorClassFrequency
		stateClass = class.StateMeasurement
	case bandwidthPropName:
		name = "Wi-Fi Bandwidth"
		id = "wi_fi_bandwidth"
		units = "MHz"
		deviceClass = class.SensorClassFrequency
		stateClass = class.StateMeasurement
	case strPropName:
		name = "Wi-Fi Signal Strength"
		id = "wi_fi_signal_strength"
		units = "%"
		stateClass = class.StateMeasurement
		icon = generateStrIcon(value)
	}

	return sensor.NewSensor(ctx,
		sensor.WithName(name),
		sensor.WithID(id),
		sensor.AsDiagnostic(),
		sensor.WithIcon(icon),
		sensor.WithState(generateState(prop, value)),
		sensor.WithDeviceClass(deviceClass),
		sensor.WithStateClass(stateClass),
		sensor.WithUnits(units),
	)
}

func getWifiSensors(ctx context.Context, bus *dbusx.Bus, apPath string) []models.Entity {
	sensors := make([]models.Entity, 0, len(apPropList))

	for _, prop := range apPropList {
		value, err := dbusx.NewProperty[any](bus, apPath, dBusNMObj, accessPointInterface+"."+prop).Get()
		if err != nil {
			slogctx.FromCtx(ctx).Debug("Could not retrieve access point property.",
				slog.String("prop", prop),
				slog.Any("error", err))

			continue
		}
		sensors = append(sensors, newWifiSensor(ctx, prop, value))
	}

	return sensors
}

func generateState(prop string, value any) any {
	switch prop {
	case ssidPropName:
		if value, ok := value.([]uint8); ok {
			return string(value)
		}

		return unknownState
	case hwAddrPropName:
		if value, ok := value.(string); ok {
			return value
		}

		return unknownState
	case freqPropName, maxBitRatePropName, bandwidthPropName:
		if value, ok := value.(uint32); ok {
			return value
		}

		return unknownState
	case strPropName:
		if value, ok := value.(uint8); ok {
			return value
		}

		return unknownState
	default:
		return unknownState
	}
}

//nolint:mnd // not useful.
func generateStrIcon(value any) string {
	str, ok := value.(uint8)

	switch {
	case !ok:
		return "mdi:wifi-strength-alert-outline"
	case str <= 25:
		return "mdi:wifi-strength-1"
	case str > 25 && str <= 50:
		return "mdi:wifi-strength-2"
	case str > 50 && str <= 75:
		return "mdi:wifi-strength-3"
	case str > 75:
		return "mdi:wifi-strength-4"
	default:
		return "mdi:wifi-strength-alert-outline"
	}
}
