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
	"github.com/joshuar/go-hass-agent/internal/linux"
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

	newDevice := &linuxDevice{}

	// Try to fetch hostname, vendor, model from DBus. Fall back to
	// /sys/devices/virtual/dmi/id for vendor and model if DBus doesn't work.
	// Ref:
	// https://github.com/ansible/ansible/blob/devel/lib/ansible/module_utils/facts/hardware/linux.py
	newDevice.hostname = linux.GetHostname(ctx)
	newDevice.hwVendor, newDevice.hwModel = linux.GetHardwareDetails(ctx)

	// Grab everything from the /etc/os-release file.
	osrelease, err := osrelease.Read()
	if err != nil {
		log.Fatal().Caller().
			Msgf("Unable to read file /etc/os-release: %v", err)
	}
	newDevice.osRelease = osrelease

	// Use the current user's username to construct an app ID.
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal().Caller().
			Msgf("Could not retrieve current user details: %v", err.Error())
	}
	newDevice.appID = Name + "-" + currentUser.Username

	// Generate a semi-random machine ID.
	machineID, err := sd_id128.GetRandomUUID()
	if err != nil {
		log.Fatal().Caller().
			Msgf("Could not retrieve a machine ID: %v", err)
	}
	newDevice.machineID = machineID.String()

	return newDevice
}

func SetupContext(ctx context.Context) context.Context {
	deviceAPI := linux.NewDeviceAPI(ctx)
	if deviceAPI == nil {
		log.Warn().Msg("No DBus connections could be established.")
		return ctx
	} else {
		deviceCtx := linux.StoreAPIInContext(ctx, deviceAPI)
		return deviceCtx
	}
}
