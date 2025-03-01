// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package media

import (
	"context"
	"errors"
	"log/slog"

	pwmonitor "github.com/ConnorsApps/pipewire-monitor-go"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
)

const (
	webcamUsageWorkerID            = "webcam_usage_sensor"
	webcamUsageWorkerPreferencesID = webcamUsageWorkerID
)

var (
	ErrInitWebcamUsageWorker = errors.New("could not init webcam usage worker")
	ErrNewWebcamUsageSensor  = errors.New("could not create webcam usage sensor")
)

// webcamPipewireEventFilter filters the pipewire events. For webcam monitoring, we are only
// interested in events of type EventNode that have the "media.class" property
// of "Video/Source".
func webcamPipewireEventFilter(e *pwmonitor.Event) bool {
	if e.Type == pwmonitor.EventNode || e.IsRemovalEvent() {
		// Parse props.
		props, err := e.NodeProps()
		if err != nil {
			return false
		}
		// Filter for v4l2 node events.
		if props.MediaClass == "Video/Source" {
			return true
		}
	}

	return false
}

func webcamUseIcon(value bool) string {
	switch value {
	case true:
		return "mdi:webcam"
	default:
		return "mdi:webcam-off"
	}
}

func newWebcamUsageSensor(ctx context.Context, inUse bool) (*models.Entity, error) {
	// Generate sensor entity.
	webcamUseSensor, err := sensor.NewSensor(ctx,
		sensor.WithName("Webcam In Use"),
		sensor.WithID("webcam_in_use"),
		sensor.AsTypeBinarySensor(),
		// sensor.WithDeviceClass(class.Binar),
		sensor.WithIcon(webcamUseIcon(inUse)),
		sensor.WithState(inUse),
		sensor.WithDataSourceAttribute(linux.DataSrcSysfs),
	)
	if err != nil {
		return nil, errors.Join(ErrNewWebcamUsageSensor, err)
	}

	return &webcamUseSensor, nil
}

type webcamUsageWorker struct {
	prefs *preferences.CommonWorkerPrefs
	inUse bool
}

// parsePWState parses a pipewire state value into the appropriate boolean value.
func (w *webcamUsageWorker) parsePWState(state pwmonitor.State) {
	switch state {
	case pwmonitor.StateRunning, pwmonitor.StateIdle:
		w.inUse = true
	case pwmonitor.StateSuspended:
		fallthrough
	default:
		w.inUse = false
	}
}

//nolint:dupl
func (w *webcamUsageWorker) Events(ctx context.Context) (<-chan models.Entity, error) {
	outCh := make(chan models.Entity)
	pwEvents := make(chan []*pwmonitor.Event)

	go monitorPipewire(ctx, pwEvents, webcamPipewireEventFilter)

	go func() {
		defer close(outCh)
		defer close(pwEvents)

		for {
			select {
			case events := <-pwEvents:
				for _, event := range events {
					w.parsePWState(*event.Info.State)

					webcamUseSensor, err := newWebcamUsageSensor(ctx, w.inUse)
					if err != nil {
						logging.FromContext(ctx).Warn("Could not parse pipewire event for webcam usage.",
							slog.Any("error", err))
					}

					outCh <- *webcamUseSensor
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return outCh, nil
}

func (w *webcamUsageWorker) Sensors(ctx context.Context) ([]models.Entity, error) {
	webcamUsage, err := newWebcamUsageSensor(ctx, w.inUse)
	if err != nil {
		return nil, errors.Join(ErrNewWebcamUsageSensor, err)
	}

	return []models.Entity{*webcamUsage}, nil
}

func (w *webcamUsageWorker) PreferencesID() string {
	return webcamUsageWorkerPreferencesID
}

func (w *webcamUsageWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func NewWebcamUsageWorker(_ context.Context) (*linux.EventSensorWorker, error) {
	var err error

	worker := linux.NewEventSensorWorker(webcamUsageWorkerID)
	webcamUsageWorker := &webcamUsageWorker{}

	webcamUsageWorker.prefs, err = preferences.LoadWorker(webcamUsageWorker)
	if err != nil {
		return nil, errors.Join(ErrInitWebcamUsageWorker, err)
	}

	if webcamUsageWorker.prefs.IsDisabled() {
		return worker, nil
	}

	worker.EventSensorType = webcamUsageWorker

	return worker, nil
}
