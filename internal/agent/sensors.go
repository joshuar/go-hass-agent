// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver,unexported-return
package agent

import (
	"errors"
	"net"
	"time"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

const (
	versionWorkerID    = "agent_version_sensor"
	externalIPWorkerID = "external_ip_sensor" //nolint:gosec // false positive

	ExternalIPUpdateInterval       = 5 * time.Minute
	ExternalIPUpdateJitter         = 10 * time.Second
	ExternalIPUpdateRequestTimeout = 15 * time.Second
)

var ipLookupHosts = map[string]map[int]string{
	"icanhazip": {4: "https://4.icanhazip.com", 6: "https://6.icanhazip.com"},
	"ipify":     {4: "https://api.ipify.org", 6: "https://api6.ipify.org"},
}

var (
	ErrInvalidIP     = errors.New("invalid IP address")
	ErrNoLookupHosts = errors.New("no IP lookup hosts found")
)

type version string

type address struct {
	addr net.IP
}

func (v *version) Name() string { return "Go Hass Agent Version" }

func (v *version) ID() string { return "agent_version" }

func (v *version) Icon() string { return "mdi:face-agent" }

func (v *version) SensorType() types.SensorClass { return types.Sensor }

func (v *version) DeviceClass() types.DeviceClass { return 0 }

func (v *version) StateClass() types.StateClass { return 0 }

func (v *version) State() any { return preferences.AppVersion }

func (v *version) Units() string { return "" }

func (v *version) Category() string { return "diagnostic" }

func (v *version) Attributes() map[string]any { return nil }

func (a *address) Name() string {
	switch {
	case a.addr.To4() != nil:
		return "External IPv4 Address"
	case a.addr.To16() != nil:
		return "External IPv6 Address"
	default:
		return "External IP Address"
	}
}

func (a *address) ID() string {
	switch {
	case a.addr.To4() != nil:
		return "external_ipv4_address"
	case a.addr.To16() != nil:
		return "external_ipv6_address"
	default:
		return "external_ip_address"
	}
}

func (a *address) Icon() string {
	switch {
	case a.addr.To4() != nil:
		return "mdi:numeric-4-box-outline"
	case a.addr.To16() != nil:
		return "mdi:numeric-6-box-outline"
	default:
		return "mdi:ip"
	}
}

func (a *address) SensorType() types.SensorClass { return types.Sensor }

func (a *address) DeviceClass() types.DeviceClass { return 0 }

func (a *address) StateClass() types.StateClass { return 0 }

func (a *address) State() any { return a.addr.String() }

func (a *address) Units() string { return "" }

func (a *address) Category() string { return "diagnostic" }

func (a *address) Attributes() map[string]any {
	attributes := make(map[string]any)
	attributes["last_updated"] = time.Now().Format(time.RFC3339)

	return attributes
}
