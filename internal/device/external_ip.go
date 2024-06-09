// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver,unexported-return
package device

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
)

const (
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

func (a *address) SensorType() types.SensorClass { return types.Sensor }

func (a *address) DeviceClass() types.DeviceClass { return 0 }

func (a *address) StateClass() types.StateClass { return 0 }

func (a *address) State() any { return a.addr.String() }

func (a *address) Units() string { return "" }

func (a *address) Category() string { return "diagnostic" }

func (a *address) Attributes() any {
	now := time.Now()

	return &struct {
		LastUpdated string `json:"last_updated"`
	}{
		LastUpdated: now.Format(time.RFC3339),
	}
}

func lookupExternalIPs(client *resty.Client, ver int) (*address, error) {
	for host, addr := range ipLookupHosts {
		log.Trace().Msgf("Trying to find external IP addresses with %s", host)
		log.Trace().
			Str("method", "GET").
			Str("url", addr[ver]).
			Time("sent_at", time.Now()).
			Msg("Fetching external IP.")

		resp, err := client.R().Get(addr[ver])
		if err != nil || resp.IsError() {
			return nil, fmt.Errorf("could not retrieve external v%d address with %s: %w", ver, addr[ver], err)
		}

		log.Trace().Err(err).
			Int("statuscode", resp.StatusCode()).
			Str("status", resp.Status()).
			Str("protocol", resp.Proto()).
			Dur("time", resp.Time()).
			Time("received_at", resp.ReceivedAt()).
			Str("body", string(resp.Body())).Msg("Response received.")

		cleanResp := strings.TrimSpace(string(resp.Body()))

		a := net.ParseIP(cleanResp)
		if a == nil {
			return nil, ErrInvalidIP
		}

		return &address{addr: a}, nil
	}

	return nil, ErrNoLookupHosts
}

type externalIPWorker struct {
	client *resty.Client
}

func (w *externalIPWorker) Name() string { return "External IP Address Sensor" }

func (w *externalIPWorker) Description() string {
	return "Sensor for the external IP addresses of the device."
}

//nolint:mnd
func (w *externalIPWorker) Sensors(_ context.Context) ([]sensor.Details, error) {
	sensors := make([]sensor.Details, 0, 2)

	for _, ver := range []int{4, 6} {
		ipAddr, err := lookupExternalIPs(w.client, ver)
		if err != nil || ipAddr == nil {
			log.Trace().Err(err).Msg("IP lookup failed.")

			continue
		}

		sensors = append(sensors, ipAddr)
	}

	return sensors, nil
}

func (w *externalIPWorker) Updates(ctx context.Context) (<-chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)

	updater := func(_ time.Duration) {
		sensors, err := w.Sensors(ctx)
		if err != nil {
			log.Debug().Err(err).Msg("Could not get IP update.")
		}

		for _, s := range sensors {
			sensorCh <- s
		}
	}
	go func() {
		defer close(sensorCh)
		helpers.PollSensors(ctx, updater, ExternalIPUpdateInterval, ExternalIPUpdateJitter)
	}()

	return sensorCh, nil
}

func NewExternalIPUpdaterWorker() *externalIPWorker {
	return &externalIPWorker{
		client: resty.New().SetTimeout(ExternalIPUpdateRequestTimeout),
	}
}
