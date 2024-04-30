// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package device

import (
	"context"
	"net"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
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

func (a *address) SensorType() types.SensorClass {
	return types.Sensor
}

func (a *address) DeviceClass() types.DeviceClass {
	return 0
}

func (a *address) StateClass() types.StateClass {
	return 0
}

func (a *address) State() any {
	return a.addr.String()
}

func (a *address) Units() string {
	return ""
}

func (a *address) Category() string {
	return "diagnostic"
}

func (a *address) Attributes() any {
	now := time.Now()
	return &struct {
		LastUpdated string `json:"Last Updated"`
	}{
		LastUpdated: now.Format(time.RFC3339),
	}
}

func lookupExternalIPs(client *resty.Client, ver int) *address {
	for host, addr := range ipLookupHosts {
		log.Trace().Msgf("Trying to find external IP addresses with %s", host)
		log.Trace().
			Str("method", "GET").
			Str("url", addr[ver]).
			Time("sent_at", time.Now()).
			Msg("Fetching external IP.")
		resp, err := client.R().Get(addr[ver])
		if err != nil {
			log.Trace().Err(err).
				Msgf("Error retrieving external v%d address with %s.", ver, addr[ver])
		} else {
			log.Trace().Err(err).
				Int("statuscode", resp.StatusCode()).
				Str("status", resp.Status()).
				Str("protocol", resp.Proto()).
				Dur("time", resp.Time()).
				Time("received_at", resp.ReceivedAt()).
				Str("body", string(resp.Body())).Msg("Response received.")
			if a := net.ParseIP(string(resp.Body())); a != nil {
				return &address{addr: a}
			}
		}
	}
	return nil
}

func ExternalIPUpdater(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details)
	client := resty.New().SetTimeout(15 * time.Second)

	updateExternalIP := func(_ time.Duration) {
		for _, ver := range []int{4, 6} {
			if ip := lookupExternalIPs(client, ver); ip != nil {
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
