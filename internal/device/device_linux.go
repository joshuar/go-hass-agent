// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package device

import (
	"context"
	"os"
	"os/user"
	"strings"

	"git.lukeshu.com/go/libsystemd/sd_id128"
	"github.com/acobaugh/osrelease"
	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
)

const (
	Name    = "go-hass-agent"
	Version = "0.0.3"
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

	// Try to fetch hostname, vendor, model from DBus. Fall back to
	// /sys/devices/virtual/dmi/id for vendor and model if DBus doesn't work.
	// Ref:
	// https://github.com/ansible/ansible/blob/devel/lib/ansible/module_utils/facts/hardware/linux.py
	deviceAPI, deviceAPIExists := FromContext(ctx)
	if !deviceAPIExists {
		log.Debug().Caller().
			Msg("Could not connect to DBus to monitor network.")
		return nil
	}
	var dBusDest = "org.freedesktop.hostname1"
	var dBusPath = "/org/freedesktop/hostname1"
	hostnameFromDBus, err := deviceAPI.GetDBusProp(systemBus,
		dBusDest,
		dbus.ObjectPath(dBusPath),
		dBusDest+".Hostname")
	if err != nil {
		device.hostname = "localhost"
	} else {
		device.hostname = string(variantToValue[[]uint8](hostnameFromDBus))
	}
	hwVendorFromDBus, err := deviceAPI.GetDBusProp(systemBus,
		dBusDest,
		dbus.ObjectPath(dBusPath),
		dBusDest+".HardwareVendor")
	if err != nil {
		hwVendor, err := os.ReadFile("/sys/devices/virtual/dmi/id/board_vendor")
		if err != nil {
			device.hwVendor = "Unknown Vendor"
		} else {
			device.hwVendor = strings.TrimSpace(string(hwVendor))
		}
	} else {
		device.hwVendor = string(variantToValue[[]uint8](hwVendorFromDBus))
	}
	hwModelFromDBus, err := deviceAPI.GetDBusProp(systemBus,
		dBusDest,
		dbus.ObjectPath(dBusPath),
		dBusDest+".HardwareModel")
	if err != nil {
		hwModel, err := os.ReadFile("/sys/devices/virtual/dmi/id/product_name")
		if err != nil {
			device.hwModel = "Unknown Vendor"
		} else {
			device.hwModel = strings.TrimSpace(string(hwModel))
		}
	} else {
		device.hwModel = string(variantToValue[[]uint8](hwModelFromDBus))
	}

	// Grab everything from the /etc/os-release file.
	osrelease, err := osrelease.Read()
	if err != nil {
		log.Fatal().Caller().
			Msgf("Unable to read file /etc/os-release: %v", err)
	}
	device.osRelease = osrelease

	// Use the current user's username to construct an app ID.
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal().Caller().
			Msgf("Could not retrieve current user details: %v", err.Error())
	}
	device.appID = Name + "-" + currentUser.Username

	// Generate a semi-random machine ID.
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
	sensorInfo.Add("Location", LocationUpdater)
	sensorInfo.Add("Battery", BatteryUpdater)
	sensorInfo.Add("Apps", AppUpdater)
	sensorInfo.Add("Network", NetworkUpdater)
	sensorInfo.Add("Power", PowerUpater)
	sensorInfo.Add("ExternalIP", ExternalIPUpdater)
	// Add each SensorUpdater function here...
	return sensorInfo
}
