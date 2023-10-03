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
}

// Setup returns a new Context that contains the D-Bus API.
func (l *LinuxDevice) Setup(ctx context.Context) context.Context {
	return context.WithValue(ctx, linuxCtxKey, newDBusAPI(ctx))
}

func (l *LinuxDevice) AppName() string {
	return l.appName
}

func (l *LinuxDevice) AppVersion() string {
	return l.appVersion
}

func (l *LinuxDevice) AppID() string {
	// Use the current user's username to construct an app ID.
	currentUser, err := user.Current()
	if err != nil {
		log.Warn().Err(err).
			Msg("Could not retrieve current user details.")
		return l.appName + "-unknown"
	}
	return l.appName + "-" + currentUser.Username
}

func (l *LinuxDevice) DeviceName() string {
	shortHostname, _, _ := strings.Cut(getHostname(), ".")
	return shortHostname
}

func (l *LinuxDevice) DeviceID() string {
	machineID, err := host.HostID()
	if err != nil {
		log.Warn().Err(err).
			Msg("Could not retrieve a machine ID")
		return "unknown"
	}
	return machineID
}

func (l *LinuxDevice) Manufacturer() string {
	return getHWVendor()
}

func (l *LinuxDevice) Model() string {
	return getHWModel()
}

func (l *LinuxDevice) OsName() string {
	_, osRelease, _, err := host.PlatformInformation()
	if err != nil {
		log.Warn().Err(err).
			Msg("Could not retrieve distribution details.")
		return "Unknown OS"
	}
	return osRelease
}

func (l *LinuxDevice) OsVersion() string {
	_, _, osVersion, err := host.PlatformInformation()
	if err != nil {
		log.Warn().Err(err).
			Msg("Could not retrieve version details.")
		return "Unknown Version"
	}
	return osVersion
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

func NewDevice(name, version string) *LinuxDevice {
	return &LinuxDevice{
		appName:    name,
		appVersion: version,
	}
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

// getHWVendor will try to retrieve the vendor from the sysfs filesystem. It
// will return "Unknown Vendor" if unsuccessful.
// Reference: https://github.com/ansible/ansible/blob/devel/lib/ansible/module_utils/facts/hardware/linux.py
func getHWVendor() string {
	hwVendor, err := os.ReadFile("/sys/devices/virtual/dmi/id/board_vendor")
	if err != nil {
		return "Unknown Vendor"
	} else {
		return strings.TrimSpace(string(hwVendor))
	}
}

// getHWModel will try to retrieve the hardware model from the sysfs filesystem. It
// will return "Unknown Model" if unsuccessful.
// Reference: https://github.com/ansible/ansible/blob/devel/lib/ansible/module_utils/facts/hardware/linux.py
func getHWModel() string {
	hwModel, err := os.ReadFile("/sys/devices/virtual/dmi/id/product_name")
	if err != nil {
		return "Unknown Model"
	} else {
		return strings.TrimSpace(string(hwModel))
	}
}
