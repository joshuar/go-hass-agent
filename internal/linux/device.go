// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"encoding/json"
	"os"
	"os/user"
	"strings"
	"sync"

	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/host"
)

type dBusAPI struct {
	dbus map[dbusType]*Bus
	mu   sync.Mutex
}

func newDBusAPI(ctx context.Context) *dBusAPI {
	a := &dBusAPI{}
	a.dbus = make(map[dbusType]*Bus)
	a.mu.Lock()
	a.dbus[SessionBus] = NewBus(ctx, SessionBus)
	a.dbus[SystemBus] = NewBus(ctx, SystemBus)
	a.mu.Unlock()
	return a
}

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type key int

// linuxCtxKey is the key for dbusAPI values in Contexts. It is unexported;
// clients use Setup and getBus instead of using this key directly.
var linuxCtxKey key

// getBus retrieves the D-Bus API object from the context
func getBus(ctx context.Context, e dbusType) (*Bus, bool) {
	b, ok := ctx.Value(linuxCtxKey).(*dBusAPI)
	if !ok {
		return nil, false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.dbus[e], true
}

type LinuxDevice struct {
	appName    string
	appVersion string
	appID      string
	hostname   string
	hwVendor   string
	hwModel    string
	osRelease  string
	osVersion  string
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
	return l.osRelease
}

func (l *LinuxDevice) OsVersion() string {
	return l.osVersion
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
	return json.Marshal(&api.RegistrationRequest{
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

// Setup returns a new Context that contains the D-Bus API.
func (l *LinuxDevice) Setup(ctx context.Context) context.Context {
	return context.WithValue(ctx, linuxCtxKey, newDBusAPI(ctx))
}

func NewDevice(name, version string) *LinuxDevice {
	device := &LinuxDevice{
		appName:    name,
		appVersion: version,
	}
	var err error

	_, device.osRelease, device.osVersion, err = host.PlatformInformation()
	if err != nil {
		log.Fatal().Caller().
			Msgf("Could not retrieve distribution details: %v", err.Error())
	}

	device.hostname = getHostname()
	device.hwVendor, device.hwModel = getHardwareDetails()

	// Use the current user's username to construct an app ID.
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal().Caller().
			Msgf("Could not retrieve current user details: %v", err.Error())
	}
	device.appID = name + "-" + currentUser.Username

	// Generate a semi-random machine ID.
	device.machineID, err = host.HostID()
	if err != nil {
		log.Fatal().Caller().
			Msgf("Could not retrieve a machine ID: %v", err)
	}

	return device
}

// getHardwareDetails will try to read the vendor and model details them from
// the /sys filesystem. If that fails, it returns empty strings for these values
// https://github.com/ansible/ansible/blob/devel/lib/ansible/module_utils/facts/hardware/linux.py
func getHardwareDetails() (string, string) {
	var vendor, model string
	hwVendor, err := os.ReadFile("/sys/devices/virtual/dmi/id/board_vendor")
	if err != nil {
		vendor = "Unknown Vendor"
	} else {
		vendor = strings.TrimSpace(string(hwVendor))
	}
	hwModel, err := os.ReadFile("/sys/devices/virtual/dmi/id/product_name")
	if err != nil {
		model = "Unknown Vendor"
	} else {
		model = strings.TrimSpace(string(hwModel))
	}
	return vendor, model
}

// getHostname retrieves the hostname of the device running the agent, or
// localhost if that doesn't work
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Warn().Err(err).Msg("Could not retrieve hostname.")
		return "localhost"
	}
	return hostname
}
