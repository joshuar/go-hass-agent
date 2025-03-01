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
	micUsageWorkerID = "microphone_usage_sensor"
)

var (
	ErrInitMicUsageWorker = errors.New("could not init mic usage worker")
	ErrNewMicUsageSensor  = errors.New("could not create mic usage sensor")
)

// micPipewireEventFilter filters the pipewire events. For mic monitoring, we are only
// interested in events of type EventNode that have the audio source media type.
func micPipewireEventFilter(e *pwmonitor.Event) bool {
	if e.Type == pwmonitor.EventNode || e.IsRemovalEvent() {
		// Parse props.
		props, err := e.NodeProps()
		if err != nil {
			return false
		}
		// Filter for audio stream events.
		return props.MediaClass == pwmonitor.MediaAudioSource
	}

	return false
}

func micUseIcon(value bool) string {
	switch value {
	case true:
		return "mdi:microphone"
	default:
		return "mdi:microphone-off"
	}
}

func newMicUsageSensor(ctx context.Context, inUse bool) (*models.Entity, error) {
	// Generate sensor entity.
	micUseSensor, err := sensor.NewSensor(ctx,
		sensor.WithName("Microphone In Use"),
		sensor.WithID("microphone_in_use"),
		sensor.AsTypeBinarySensor(),
		// sensor.WithDeviceClass(class.Binar),
		sensor.WithIcon(micUseIcon(inUse)),
		sensor.WithState(inUse),
		sensor.WithDataSourceAttribute(linux.DataSrcSysfs),
	)
	if err != nil {
		return nil, errors.Join(ErrNewMicUsageSensor, err)
	}

	return &micUseSensor, nil
}

type micUsageWorker struct {
	prefs *preferences.CommonWorkerPrefs
	inUse bool
}

// parsePWState parses a pipewire state value into the appropriate boolean value.
func (w *micUsageWorker) parsePWState(state pwmonitor.State) {
	switch state {
	case pwmonitor.StateRunning:
		w.inUse = true
	case pwmonitor.StateIdle, pwmonitor.StateSuspended:
		fallthrough
	default:
		w.inUse = false
	}
}

//nolint:dupl
func (w *micUsageWorker) Events(ctx context.Context) (<-chan models.Entity, error) {
	outCh := make(chan models.Entity)
	pwEvents := make(chan []*pwmonitor.Event)

	go monitorPipewire(ctx, pwEvents, micPipewireEventFilter)

	go func() {
		defer close(outCh)
		defer close(pwEvents)

		for {
			select {
			case events := <-pwEvents:
				for _, event := range events {
					w.parsePWState(*event.Info.State)

					micUseSensor, err := newMicUsageSensor(ctx, w.inUse)
					if err != nil {
						logging.FromContext(ctx).Warn("Could not parse pipewire event for mic usage.",
							slog.Any("error", err))
					}

					outCh <- *micUseSensor
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return outCh, nil
}

func (w *micUsageWorker) Sensors(ctx context.Context) ([]models.Entity, error) {
	micUsage, err := newMicUsageSensor(ctx, w.inUse)
	if err != nil {
		return nil, errors.Join(ErrNewMicUsageSensor, err)
	}

	return []models.Entity{*micUsage}, nil
}

func (w *micUsageWorker) PreferencesID() string {
	return micUsagePrefID
}

func (w *micUsageWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func NewMicUsageWorker(_ context.Context) (*linux.EventSensorWorker, error) {
	var err error

	worker := linux.NewEventSensorWorker(micUsageWorkerID)
	micUsageWorker := &micUsageWorker{}

	micUsageWorker.prefs, err = preferences.LoadWorker(micUsageWorker)
	if err != nil {
		return nil, errors.Join(ErrInitMicUsageWorker, err)
	}

	if micUsageWorker.prefs.IsDisabled() {
		return worker, nil
	}

	worker.EventSensorType = micUsageWorker

	return worker, nil
}
