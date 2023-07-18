// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"encoding/json"
	"os/user"
	"strings"

	"git.lukeshu.com/go/libsystemd/sd_id128"
	"github.com/acobaugh/osrelease"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

type LinuxDevice struct {
	appName    string
	appVersion string
	appID      string
	hostname   string
	hwVendor   string
	hwModel    string
	osRelease  map[string]string
	machineID  string
}

func (l *LinuxDevice) AppName() string {
	return l.appName
}

func (l *LinuxDevice) AppVersion() string {
	return l.appVersion
}

func (l *LinuxDevice) AppID() string {
	return l.appID
}

func (l *LinuxDevice) DeviceName() string {
	shortHostname, _, _ := strings.Cut(l.hostname, ".")
	return shortHostname
}

func (l *LinuxDevice) DeviceID() string {
	return l.machineID
}

func (l *LinuxDevice) Manufacturer() string {
	return l.hwVendor
}

func (l *LinuxDevice) Model() string {
	return l.hwModel
}

func (l *LinuxDevice) OsName() string {
	return l.osRelease["PRETTY_NAME"]
}

func (l *LinuxDevice) OsVersion() string {
	return l.osRelease["VERSION_ID"]
}

func (l *LinuxDevice) SupportsEncryption() bool {
	return false
}

func (l *LinuxDevice) AppData() interface{} {
	return &struct {
		PushWebsocket bool `json:"push_websocket_channel"`
	}{
		PushWebsocket: true,
	}
}

func (l *LinuxDevice) MarshalJSON() ([]byte, error) {
	return json.Marshal(&hass.RegistrationRequest{
		DeviceID:           l.DeviceID(),
		AppID:              l.AppID(),
		AppName:            l.AppName(),
		AppVersion:         l.AppVersion(),
		DeviceName:         l.DeviceName(),
		Manufacturer:       l.Manufacturer(),
		Model:              l.Model(),
		OsName:             l.OsName(),
		OsVersion:          l.OsVersion(),
		SupportsEncryption: l.SupportsEncryption(),
		AppData:            l.AppData(),
	})
}

func NewDevice(ctx context.Context, name string, version string) *LinuxDevice {
	newDevice := &LinuxDevice{
		appName:    name,
		appVersion: version,
	}

	// Try to fetch hostname, vendor, model from DBus. Fall back to
	// /sys/devices/virtual/dmi/id for vendor and model if DBus doesn't work.
	// Ref:
	// https://github.com/ansible/ansible/blob/devel/lib/ansible/module_utils/facts/hardware/linux.py
	newDevice.hostname = GetHostname(ctx)
	newDevice.hwVendor, newDevice.hwModel = GetHardwareDetails(ctx)

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
	newDevice.appID = name + "-" + currentUser.Username

	// Generate a semi-random machine ID.
	machineID, err := sd_id128.GetRandomUUID()
	if err != nil {
		log.Fatal().Caller().
			Msgf("Could not retrieve a machine ID: %v", err)
	}
	newDevice.machineID = machineID.String()

	return newDevice
}

// func SetupContext(ctx context.Context) context.Context {
// 	deviceAPI := api.NewDeviceAPI(ctx)
// 	if deviceAPI == nil {
// 		log.Warn().Msg("No DBus connections could be established.")
// 		return ctx
// 	} else {
// 		deviceCtx := api.StoreAPIInContext(ctx, deviceAPI)
// 		return deviceCtx
// 	}
// }
