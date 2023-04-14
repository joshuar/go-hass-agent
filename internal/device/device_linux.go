// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package device

import (
	"context"
	"os/user"
	"strings"

	"git.lukeshu.com/go/libsystemd/sd_id128"
	"github.com/acobaugh/osrelease"
	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
)

const (
	Name    = "go-hass-agent"
	Version = "0.0.1"
)

type linuxDevice struct {
	hostname  string
	hwVendor  string
	hwModel   string
	osRelease map[string]string
	appID     string
	machineID string
}

func (l *linuxDevice) AppName() string {
	return Name
}

func (l *linuxDevice) AppVersion() string {
	return Version
}

func (l *linuxDevice) AppID() string {
	return l.appID
}

func (l *linuxDevice) DeviceName() string {
	shortHostname, _, _ := strings.Cut(l.hostname, ".")
	return shortHostname
}

func (l *linuxDevice) DeviceID() string {
	return l.machineID
}

func (l *linuxDevice) Manufacturer() string {
	return l.hwVendor
}

func (l *linuxDevice) Model() string {
	return l.hwModel
}

func (l *linuxDevice) OsName() string {
	return l.osRelease["PRETTY_NAME"]
}

func (l *linuxDevice) OsVersion() string {
	return l.osRelease["VERSION_ID"]
}

func (l *linuxDevice) SupportsEncryption() bool {
	return false
}

func (l *linuxDevice) AppData() interface{} {
	return &struct {
		PushWebsocket bool `json:"push_websocket_channel"`
	}{
		PushWebsocket: true,
	}
}

func NewDevice(ctx context.Context) *linuxDevice {

	device := &linuxDevice{}

	deviceAPI, deviceAPIExists := FromContext(ctx)
	if !deviceAPIExists {
		log.Debug().Caller().
			Msg("Could not connect to DBus to monitor network.")
		return nil
	}

	var dBusDest = "org.freedesktop.hostname1"
	var dBusPath = "/org/freedesktop/hostname1"

	hostnameFromDBus := deviceAPI.GetDBusProp(systemBus,
		dBusDest,
		dbus.ObjectPath(dBusPath),
		dBusDest+".Hostname")
	if hostname := hostnameFromDBus.Value().(string); hostname != "" {
		device.hostname = hostname
	} else {
		device.hostname = "localhost"
	}

	hwVendorFromDBus := deviceAPI.GetDBusProp(systemBus,
		dBusDest,
		dbus.ObjectPath(dBusPath),
		dBusDest+".HardwareVendor")
	if vendor := hwVendorFromDBus.Value().(string); vendor != "" {
		device.hwVendor = vendor
	} else {
		device.hwVendor = "Unknown Vendor"
	}

	hwModelFromDBus := deviceAPI.GetDBusProp(systemBus,
		dBusDest,
		dbus.ObjectPath(dBusPath),
		dBusDest+".HardwareModel")
	if model := hwModelFromDBus.Value().(string); model != "" {
		device.hwModel = model
	} else {
		device.hwModel = "Unknown Vendor"
	}

	osrelease, err := osrelease.Read()
	if err != nil {
		log.Fatal().Caller().
			Msgf("Unable to read file /etc/os-release: %v", err)
	}
	device.osRelease = osrelease

	currentUser, err := user.Current()
	if err != nil {
		log.Fatal().Caller().
			Msgf("Could not retrieve current user details: %v", err.Error())
	}
	device.appID = Name + "-" + currentUser.Username

	machineID, err := sd_id128.GetRandomUUID()
	if err != nil {
		log.Fatal().Caller().
			Msgf("Could not retrieve a machine ID: %v", err)
	}
	device.machineID = machineID.String()

	return device
}

type deviceAPI struct {
	dBusSystem, dBusSession *bus
	WatchEvents             chan *DBusWatchRequest
}

func SetupContext(ctx context.Context) context.Context {
	deviceAPI := &deviceAPI{
		dBusSystem:  NewBus(ctx, systemBus),
		dBusSession: NewBus(ctx, sessionBus),
		WatchEvents: make(chan *DBusWatchRequest),
	}
	if deviceAPI.dBusSession == nil && deviceAPI.dBusSystem == nil {
		log.Warn().Msg("No DBus connections could be established.")
		return ctx
	} else {
		go deviceAPI.monitorDBus(ctx)
		deviceCtx := NewContext(ctx, deviceAPI)
		return deviceCtx
	}
}

func SetupSensors() *SensorInfo {
	sensorInfo := NewSensorInfo()
	sensorInfo.Add("Battery", BatteryUpdater)
	sensorInfo.Add("Apps", AppUpdater)
	sensorInfo.Add("Network", NetworkUpdater)
	// Add each SensorUpdater function here...
	return sensorInfo
}
