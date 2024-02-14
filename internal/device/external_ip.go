// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package device

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/carlmjohnson/requests"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
)

var ipLookupHosts = map[string]map[int]string{
	"icanhazip": {4: "https://4.icanhazip.com", 6: "https://6.icanhazip.com"},
	"ipify":     {4: "https://api.ipify.org", 6: "https://api6.ipify.org"},
}

type address struct {
	addr net.IP
}

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

func (a *address) SensorType() sensor.SensorType {
	return sensor.TypeSensor
}

func (a *address) DeviceClass() sensor.SensorDeviceClass {
	return 0
}

func (a *address) StateClass() sensor.SensorStateClass {
	return 0
}

func (a *address) State() interface{} {
	return a.addr.String()
}

func (a *address) Units() string {
	return ""
}

func (a *address) Category() string {
	return "diagnostic"
}

func (a *address) Attributes() interface{} {
	now := time.Now()
	return &struct {
		LastUpdated string `json:"Last Updated"`
	}{
		LastUpdated: now.Format(time.RFC3339),
	}
}

func lookupExternalIPs(ctx context.Context, ver int) chan *address {
	addrCh := make(chan *address, 1)
	defer close(addrCh)
	for host, addr := range ipLookupHosts {
		log.Trace().Msgf("Trying to find external IP addresses with %s", host)
		var s string
		err := requests.
			URL(addr[ver]).
			ToString(&s).
			Fetch(ctx)
		log.Trace().Msgf("Fetching v%d address from %s", ver, addr[ver])
		if err != nil {
			if !errors.Is(err, requests.ErrTransport) {
				log.Warn().Err(err).
					Msgf("Error retrieving external v%d address with %s.", ver, addr[ver])
			}
		} else {
			s = strings.TrimSpace(s)
			if a := net.ParseIP(s); a != nil {
				log.Trace().Msgf("Found address %s with %s.", a.String(), addr[ver])
				addrCh <- &address{addr: a}
				return addrCh
			}
		}
	}
	return addrCh
}

func ExternalIPUpdater(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details, 1)
	updateExternalIP := func(_ time.Duration) {
		requestCtx, cancel := context.WithTimeout(ctx, time.Second*15)
		defer cancel()
		for _, ver := range []int{4, 6} {
			ip := <-lookupExternalIPs(requestCtx, ver)
			if ip != nil {
				sensorCh <- ip
			}
		}
	}
	go helpers.PollSensors(ctx, updateExternalIP, 5*time.Minute, 30*time.Second)
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
	}()
	return sensorCh
}
