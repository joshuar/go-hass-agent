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
	"github.com/joshuar/go-hass-agent/internal/logging"
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

type sensorWorker struct {
	object  Worker
	started bool
}

type VersionWorker struct {
	version
}

type ExternalIPWorker struct {
	client     *resty.Client
	logger     *slog.Logger
	cancelFunc context.CancelFunc
}

type deviceController struct {
	sensorWorkers map[string]*sensorWorker
	logger        *slog.Logger
}

func (w *deviceController) ActiveWorkers() []string {
	activeWorkers := make([]string, 0, len(w.sensorWorkers))

	for id, worker := range w.sensorWorkers {
		if worker.started {
			activeWorkers = append(activeWorkers, id)
		}
	}

	return activeWorkers
}

func (w *deviceController) InactiveWorkers() []string {
	inactiveWorkers := make([]string, 0, len(w.sensorWorkers))

	for id, worker := range w.sensorWorkers {
		if !worker.started {
			inactiveWorkers = append(inactiveWorkers, id)
		}
	}

	return inactiveWorkers
}

func (w *deviceController) Start(ctx context.Context, name string) (<-chan sensor.Details, error) {
	worker, exists := w.sensorWorkers[name]
	if !exists {
		return nil, ErrUnknownWorker
	}

	if worker.started {
		return nil, ErrWorkerAlreadyStarted
	}

	workerCh, err := w.sensorWorkers[name].object.Updates(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start worker: %w", err)
	}

	w.sensorWorkers[name].started = true

	return workerCh, nil
}

func (w *deviceController) Stop(name string) error {
	// Check if the given worker ID exists.
	worker, exists := w.sensorWorkers[name]
	if !exists {
		return ErrUnknownWorker
	}
	// Stop the worker. Report any errors.
	if err := worker.object.Stop(); err != nil {
		return fmt.Errorf("error stopping worker: %w", err)
	}

	return nil
}

func (w *deviceController) StartAll(ctx context.Context) (<-chan sensor.Details, error) {
	outCh := make([]<-chan sensor.Details, 0, len(allworkers))

	var errs error

	for id := range w.sensorWorkers {
		workerCh, err := w.Start(ctx, id)
		if err != nil {
			errs = errors.Join(errs, err)

			continue
		}

		outCh = append(outCh, workerCh)
	}

	return mergeCh(ctx, outCh...), errs
}

func (w *deviceController) StopAll() error {
	var errs error

	for id := range w.sensorWorkers {
		if err := w.Stop(id); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

func (agent *Agent) newDeviceController(ctx context.Context) SensorController {
	var worker Worker

	controller := &deviceController{
		sensorWorkers: make(map[string]*sensorWorker),
		logger:        logging.FromContext(ctx).With(slog.Group("device")),
	}

	// Set up sensor workers.
	worker = agent.newVersionWorker()
	controller.sensorWorkers[worker.ID()] = &sensorWorker{object: worker, started: false}
	worker = agent.newExternalIPUpdaterWorker(ctx)
	controller.sensorWorkers[worker.ID()] = &sensorWorker{object: worker, started: false}

	return controller
}

func (w *VersionWorker) ID() string { return versionWorkerID }

func (w *VersionWorker) Stop() error { return nil }

func (w *VersionWorker) Sensors(_ context.Context) ([]sensor.Details, error) {
	return []sensor.Details{new(version)}, nil
}

func (w *VersionWorker) Updates(ctx context.Context) (<-chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)

	sensors, err := w.Sensors(ctx)
	if err != nil {
		close(sensorCh)

		return sensorCh, fmt.Errorf("unable to retrieve version info: %w", err)
	}

	go func() {
		defer close(sensorCh)
		sensorCh <- sensors[0]
	}()

	return sensorCh, nil
}

func (agent *Agent) newVersionWorker() *VersionWorker {
	if agent.prefs != nil {
		return &VersionWorker{version: version(agent.prefs.Version)}
	}

	return &VersionWorker{version: version("unknown")}
}

// ID returns the unique string to represent this worker and its sensors.
func (w *ExternalIPWorker) ID() string { return externalIPWorkerID }

// Stop will stop any processing of sensors controlled by this worker.
func (w *ExternalIPWorker) Stop() error {
	w.cancelFunc()

	return nil
}

//nolint:mnd
func (w *ExternalIPWorker) Sensors(ctx context.Context) ([]sensor.Details, error) {
	sensors := make([]sensor.Details, 0, 2)

	for _, ver := range []int{4, 6} {
		ipAddr, err := w.lookupExternalIPs(ctx, w.client, ver)
		if err != nil || ipAddr == nil {
			w.logger.Log(ctx, logging.LevelTrace, "Looking up external IP failed.", slog.Any("error", err))

			continue
		}

		sensors = append(sensors, ipAddr)
	}

	return sensors, nil
}

func (w *ExternalIPWorker) Updates(ctx context.Context) (<-chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)

	// Create a new context for the updates scope.
	updatesCtx, cancelFunc := context.WithCancel(ctx)
	// Save the context cancelFunc in the worker to be used as part of its
	// Stop() method.
	w.cancelFunc = cancelFunc

	updater := func(_ time.Duration) {
		sensors, err := w.Sensors(updatesCtx)
		if err != nil {
			w.logger.Debug("Could not get external IP.", slog.Any("error", err))
		}

		for _, s := range sensors {
			sensorCh <- s
		}
	}
	go func() {
		defer close(sensorCh)
		helpers.PollSensors(updatesCtx, updater, ExternalIPUpdateInterval, ExternalIPUpdateJitter)
	}()

	return sensorCh, nil
}

func (w *ExternalIPWorker) lookupExternalIPs(ctx context.Context, client *resty.Client, ver int) (*address, error) {
	for host, addr := range ipLookupHosts {
		w.logger.LogAttrs(ctx, logging.LevelTrace,
			"Fetching external IP.",
			slog.String("host", host),
			slog.String("method", "GET"),
			slog.String("url", addr[ver]),
			slog.Time("sent_at", time.Now()))

		resp, err := client.R().Get(addr[ver])
		if err != nil || resp.IsError() {
			return nil, fmt.Errorf("could not retrieve external v%d address with %s: %w", ver, addr[ver], err)
		}

		w.logger.LogAttrs(ctx, logging.LevelTrace,
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

func (agent *Agent) newExternalIPUpdaterWorker(ctx context.Context) *ExternalIPWorker {
	return &ExternalIPWorker{
		client: resty.New().SetTimeout(ExternalIPUpdateRequestTimeout),
		logger: logging.FromContext(ctx).With(slog.String("worker", externalIPWorkerID)),
	}
}
