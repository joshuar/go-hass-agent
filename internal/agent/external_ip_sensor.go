// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package agent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

const (
	externalIPPollInterval         = 5 * time.Minute
	externalIPJitterAmount         = 10 * time.Second
	externalIPUpdateRequestTimeout = 15 * time.Second

	externalIPWorkerID = "external_ip_sensor" //nolint:gosec // false positive
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

func (a *address) Category() string { return types.CategoryDiagnostic }

func (a *address) Attributes() map[string]any {
	attributes := make(map[string]any)
	attributes["last_updated"] = time.Now().Format(time.RFC3339)

	return attributes
}

type externalIPWorker struct {
	client *resty.Client
	doneCh chan struct{}
	logger *slog.Logger
}

// ID returns the unique string to represent this worker and its sensors.
func (w *externalIPWorker) ID() string { return externalIPWorkerID }

// Stop will stop any processing of sensors controlled by this worker.
func (w *externalIPWorker) Stop() error {
	close(w.doneCh)

	return nil
}

//nolint:mnd
func (w *externalIPWorker) Sensors(ctx context.Context) ([]sensor.Details, error) {
	sensors := make([]sensor.Details, 0, 2)

	for _, ver := range []int{4, 6} {
		ipAddr, err := w.lookupExternalIPs(ctx, ver)
		if err != nil || ipAddr == nil {
			w.logger.Log(ctx, logging.LevelTrace, "Looking up external IP failed.", slog.Any("error", err))

			continue
		}

		sensors = append(sensors, ipAddr)
	}

	return sensors, nil
}

func (w *externalIPWorker) Start(ctx context.Context) (<-chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)
	w.doneCh = make(chan struct{})

	updater := func(_ time.Duration) {
		sensors, err := w.Sensors(ctx)
		if err != nil {
			w.logger.
				With(slog.String("worker", externalIPWorkerID)).
				Debug("Could not get external IP.", slog.Any("error", err))
		}

		for _, s := range sensors {
			sensorCh <- s
		}
	}
	go func() {
		helpers.PollSensors(ctx, updater, externalIPPollInterval, externalIPJitterAmount)
	}()

	go func() {
		defer close(sensorCh)
		<-w.doneCh
	}()

	return sensorCh, nil
}

func (w *externalIPWorker) lookupExternalIPs(ctx context.Context, ver int) (*address, error) {
	for host, addr := range ipLookupHosts {
		w.logger.
			With(slog.String("worker", externalIPWorkerID)).
			LogAttrs(ctx, logging.LevelTrace,
				"Fetching external IP.",
				slog.String("host", host),
				slog.String("method", "GET"),
				slog.String("url", addr[ver]),
				slog.Time("sent_at", time.Now()))

		resp, err := w.client.R().Get(addr[ver])
		if err != nil || resp.IsError() {
			return nil, fmt.Errorf("could not retrieve external v%d address with %s: %w", ver, addr[ver], err)
		}

		w.logger.
			With(slog.String("worker", externalIPWorkerID)).
			LogAttrs(ctx, logging.LevelTrace,
				"Received external IP.",
				slog.Int("statuscode", resp.StatusCode()),
				slog.String("status", resp.Status()),
				slog.String("protocol", resp.Proto()),
				slog.Duration("time", resp.Time()),
				slog.String("body", string(resp.Body())))

		cleanResp := strings.TrimSpace(string(resp.Body()))

		a := net.ParseIP(cleanResp)
		if a == nil {
			return nil, ErrInvalidIP
		}

		return &address{addr: a}, nil
	}

	return nil, ErrNoLookupHosts
}

func newExternalIPUpdaterWorker(ctx context.Context) *externalIPWorker {
	return &externalIPWorker{
		client: resty.New().SetTimeout(externalIPUpdateRequestTimeout),
		logger: logging.FromContext(ctx).
			With(slog.String("worker", externalIPWorkerID)),
	}
}
