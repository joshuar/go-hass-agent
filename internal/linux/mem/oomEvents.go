// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package mem

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/event"
	"github.com/joshuar/go-hass-agent/internal/workers"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	oomEventsWorkerID      = "oom_events"
	oomEventsWorkerDesc    = "OOM events detection"
	oomEventsPreferencesID = prefPrefix + "oom_events"
	oomDBusPath            = "/org/freedesktop/systemd1/unit"
	unitPathPrefix         = "/org/freedesktop/systemd1/unit"
	oomEventName           = "oom_event"
)

var _ workers.EntityWorker = (*OOMEventsWorker)(nil)

var ErrInitOOMWorker = errors.New("could not init OOM worker")

type OOMEventsWorker struct {
	bus   *dbusx.Bus
	prefs *preferences.CommonWorkerPrefs
	*models.WorkerMetadata
}

//nolint:gocognit
func (w *OOMEventsWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPathNamespace(oomDBusPath),
		dbusx.MatchPropChanged(),
	).Start(ctx, w.bus)
	if err != nil {
		return nil, errors.Join(ErrInitOOMWorker,
			fmt.Errorf("unable to set-up D-Bus watch for OOM events: %w", err))
	}
	eventCh := make(chan models.Entity)

	go func() {
		defer close(eventCh)

		for {
			select {
			case <-ctx.Done():
				return
			case trigger := <-triggerCh:
				props, err := dbusx.ParsePropertiesChanged(trigger.Content)
				if err != nil {
					logging.FromContext(ctx).Debug("Could not parse changed properties for unit.", slog.Any("error", err))
					continue
				}
				// Ignore events that don't indicate a result change.
				if _, found := props.Changed["Result"]; !found {
					continue
				}

				result, err := dbusx.VariantToValue[string](props.Changed["Result"])
				if err != nil {
					logging.FromContext(ctx).Debug("Could not parse result.", slog.Any("error", err))
					continue
				}

				if result == "oom-kill" {
					// Naming is defined in
					// https://systemd.io/DESKTOP_ENVIRONMENTS/. The strings seem
					// to be percent-encoded with % replaced by _.
					processStr, err := url.PathUnescape(strings.ReplaceAll(trigger.Path, "_", "%"))
					if err != nil {
						logging.FromContext(ctx).Debug("Could not unescape process path string.", slog.Any("error", err))
					}

					// Trim the D-Bus unit path prefix.
					processStr = strings.TrimPrefix(processStr, unitPathPrefix+"/")
					// Trim any "app-" prefix.
					processStr = strings.TrimPrefix(processStr, "app-")
					// Trim the ".service" suffix.
					processStr = strings.TrimSuffix(processStr, ".service")
					// Ignore the <RANDOM> string that might be appended.
					processStr, _, _ = strings.Cut(processStr, "@")
					// Get the PID.
					//nolint:errcheck
					pid, _ := dbusx.VariantToValue[int](props.Changed["MainPID"])
					if pid == 0 {
						continue
					}
					// Send an event.
					entity, err := event.NewEvent(
						oomEventName,
						map[string]any{
							"Process": processStr,
							"PID":     pid,
						})
					if err != nil {
						logging.FromContext(ctx).Warn("Could not create OOM event.", slog.Any("error", err))
					} else {
						eventCh <- entity
					}
				}
			}
		}
	}()

	return eventCh, nil
}

func (w *OOMEventsWorker) PreferencesID() string {
	return oomEventsPreferencesID
}

func (w *OOMEventsWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func (w *OOMEventsWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func NewOOMEventsWorker(ctx context.Context) (workers.EntityWorker, error) {
	bus, ok := linux.CtxGetSessionBus(ctx)
	if !ok {
		return nil, errors.Join(ErrInitOOMWorker, linux.ErrNoSessionBus)
	}

	worker := &OOMEventsWorker{
		WorkerMetadata: models.SetWorkerMetadata(oomEventsWorkerID, oomEventsWorkerDesc),
		bus:            bus,
	}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return nil, errors.Join(ErrInitOOMWorker, err)
	}
	worker.prefs = prefs

	return worker, nil
}
