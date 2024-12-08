// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package agentsensor

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
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

const (
	externalIPPollInterval         = 5 * time.Minute
	externalIPJitterAmount         = 10 * time.Second
	externalIPUpdateRequestTimeout = 15 * time.Second

	externalIPWorkerID = "external_ip"
)

var ipLookupHosts = map[string]map[int]string{
	"icanhazip": {4: "https://4.icanhazip.com", 6: "https://6.icanhazip.com"},
	"ipify":     {4: "https://api.ipify.org", 6: "https://api6.ipify.org"},
}

var (
	ErrInvalidIP     = errors.New("invalid IP address")
	ErrNoLookupHosts = errors.New("no IP lookup hosts found")
)

func newExternalIPSensor(addr net.IP) sensor.Entity {
	var name, id, icon string

	switch {
	case addr.To4() != nil:
		name = "External IPv4 Address"
		id = "external_ipv4_address"
		icon = "mdi:numeric-4-box-outline"
	case addr.To16() != nil:
		name = "External IPv6 Address"
		id = "external_ipv6_address"
		icon = "mdi:numeric-6-box-outline"
	}

	return sensor.Entity{
		Name:     name,
		Category: types.CategoryDiagnostic,
		State: &sensor.State{
			ID:         id,
			Icon:       icon,
			EntityType: types.Sensor,
			Value:      addr.String(),
			Attributes: map[string]any{
				"last_updated": time.Now().Format(time.RFC3339),
			},
		},
	}
}

type ExternalIPWorker struct {
	client *resty.Client
	doneCh chan struct{}
	logger *slog.Logger
	prefs  *ExternalIPWorkerPrefs
}

type ExternalIPWorkerPrefs preferences.CommonWorkerPrefs

func (w *ExternalIPWorker) PreferencesID() string {
	return "external_ip_sensor"
}

func (w *ExternalIPWorker) DefaultPreferences() ExternalIPWorkerPrefs {
	return ExternalIPWorkerPrefs{}
}

func (w *ExternalIPWorker) Disabled() bool {
	return w.prefs.Disabled
}

// ID returns the unique string to represent this worker and its sensors.
func (w *ExternalIPWorker) ID() string { return externalIPWorkerID }

// Stop will stop any processing of sensors controlled by this worker.
func (w *ExternalIPWorker) Stop() error {
	close(w.doneCh)

	return nil
}

//nolint:mnd
func (w *ExternalIPWorker) Sensors(ctx context.Context) ([]sensor.Entity, error) {
	sensors := make([]sensor.Entity, 0, 2)

	for _, ver := range []int{4, 6} {
		ipAddr, err := w.lookupExternalIPs(ctx, ver)
		if err != nil || ipAddr == nil {
			w.logger.Log(ctx, logging.LevelTrace, "Looking up external IP failed.", slog.Any("error", err))

			continue
		}

		sensors = append(sensors, newExternalIPSensor(ipAddr))
	}

	return sensors, nil
}

func (w *ExternalIPWorker) Start(ctx context.Context) (<-chan sensor.Entity, error) {
	sensorCh := make(chan sensor.Entity)
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

func (w *ExternalIPWorker) lookupExternalIPs(ctx context.Context, ver int) (net.IP, error) {
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

		return a, nil
	}

	return nil, ErrNoLookupHosts
}

func NewExternalIPUpdaterWorker(ctx context.Context) *ExternalIPWorker {
	var err error

	worker := &ExternalIPWorker{
		client: resty.New().SetTimeout(externalIPUpdateRequestTimeout),
		logger: logging.FromContext(ctx).
			With(slog.String("worker", externalIPWorkerID)),
	}

	prefs, err := preferences.LoadWorker(ctx, worker)
	if err != nil {
		return nil
	}

	worker.prefs = prefs

	if worker.Disabled() {
		return nil
	}

	return worker
}
