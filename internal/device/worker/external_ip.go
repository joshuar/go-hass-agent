// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"slices"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/reugn/go-quartz/quartz"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
	"github.com/joshuar/go-hass-agent/internal/scheduler"
	"github.com/joshuar/go-hass-agent/internal/workers"
)

const (
	externalIPPollInterval         = 5 * time.Minute
	externalIPJitterAmount         = 10 * time.Second
	externalIPUpdateRequestTimeout = 15 * time.Second

	externalIPWorkerID   = "external_ip"
	externalIPWorkerDesc = "Get external IP details"
)

var ipLookupHosts = map[string]map[int]string{
	"icanhazip": {4: "https://4.icanhazip.com", 6: "https://6.icanhazip.com"},
	"ipify":     {4: "https://api.ipify.org", 6: "https://api6.ipify.org"},
}

var (
	_ quartz.Job                  = (*ExternalIP)(nil)
	_ workers.PollingEntityWorker = (*ExternalIP)(nil)
)

var (
	ErrInitExternalIPWorker = errors.New("could not init external IP worker")
	ErrInvalidIP            = errors.New("invalid IP address")
	ErrNoLookupHosts        = errors.New("no IP lookup hosts found")
)

type ExternalIP struct {
	client *resty.Client
	prefs  *preferences.CommonWorkerPrefs
	*workers.PollingEntityWorkerData
	*models.WorkerMetadata
}

func (w *ExternalIP) PreferencesID() string {
	return preferences.SensorsPrefPrefix + "agent" + preferences.PathDelim + "external_ip"
}

func (w *ExternalIP) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func (w *ExternalIP) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *ExternalIP) Execute(ctx context.Context) error {
	for ipVer := range slices.Values([]int{4, 6}) {
		ipAddr, err := w.lookupExternalIPs(ctx, ipVer)
		if err != nil || ipAddr == nil {
			slogctx.FromCtx(ctx).Log(ctx, logging.LevelTrace, "Looking up external IP failed.", slog.Any("error", err))
			continue
		}

		entity, err := newExternalIPSensor(ctx, ipAddr)
		if err != nil {
			slogctx.FromCtx(ctx).Log(ctx, logging.LevelTrace, "Sensor creation failed.", slog.Any("error", err))
			continue
		}
		w.OutCh <- entity
	}
	return nil
}

func newExternalIPSensor(ctx context.Context, addr net.IP) (models.Entity, error) {
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

	entity, err := sensor.NewSensor(ctx,
		sensor.WithName(name),
		sensor.WithID(id),
		sensor.AsDiagnostic(),
		sensor.WithIcon(icon),
		sensor.WithState(addr.String()),
		sensor.WithAttribute("last_updated", time.Now().Format(time.RFC3339)),
	)
	if err != nil {
		return entity, fmt.Errorf("could not create external IP sensor entity: %w", err)
	}

	return entity, nil
}

func (w *ExternalIP) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start disk usage worker: %w", err)
	}
	return w.OutCh, nil
}

func (w *ExternalIP) lookupExternalIPs(ctx context.Context, ver int) (net.IP, error) {
	for host, addr := range ipLookupHosts {
		slogctx.FromCtx(ctx).
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

		slogctx.FromCtx(ctx).
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

func NewExternalIPWorker(ctx context.Context) (workers.EntityWorker, error) {
	var err error

	worker := &ExternalIP{
		WorkerMetadata:          models.SetWorkerMetadata(externalIPWorkerID, externalIPWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
		client:                  resty.New().SetTimeout(externalIPUpdateRequestTimeout),
	}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return nil, errors.Join(ErrInitExternalIPWorker, err)
	}
	worker.prefs = prefs

	worker.Trigger = scheduler.NewPollTriggerWithJitter(externalIPPollInterval, externalIPJitterAmount)

	return worker, nil
}
