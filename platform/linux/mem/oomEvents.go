// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package mem

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/event"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
	"github.com/joshuar/go-hass-agent/platform/linux"
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

type OOMEventsWorker struct {
	*models.WorkerMetadata

	bus   *dbusx.Bus
	prefs *workers.CommonWorkerPrefs
}

func NewOOMEventsWorker(ctx context.Context) (workers.EntityWorker, error) {
	worker := &OOMEventsWorker{
		WorkerMetadata: models.SetWorkerMetadata(oomEventsWorkerID, oomEventsWorkerDesc),
	}

	var ok bool

	worker.bus, ok = linux.CtxGetSessionBus(ctx)
	if !ok {
		return worker, fmt.Errorf("get session bus: %w", linux.ErrNoSessionBus)
	}

	defaultPrefs := &workers.CommonWorkerPrefs{}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(oomEventsPreferencesID, defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	return worker, nil
}

//nolint:gocognit
func (w *OOMEventsWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPathNamespace(oomDBusPath),
		dbusx.MatchPropChanged(),
	).Start(ctx, w.bus)
	if err != nil {
		return nil, fmt.Errorf("watch for OOM events: %w", err)
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
					slogctx.FromCtx(ctx).Debug("Could not parse changed properties for unit.", slog.Any("error", err))
					continue
				}
				// Ignore events that don't indicate a result change.
				if _, found := props.Changed["Result"]; !found {
					continue
				}

				result, err := dbusx.VariantToValue[string](props.Changed["Result"])
				if err != nil {
					slogctx.FromCtx(ctx).Debug("Could not parse result.", slog.Any("error", err))
					continue
				}

				if result == "oom-kill" {
					// Naming is defined in
					// https://systemd.io/DESKTOP_ENVIRONMENTS/. The strings seem
					// to be percent-encoded with % replaced by _.
					processStr, err := url.PathUnescape(strings.ReplaceAll(trigger.Path, "_", "%"))
					if err != nil {
						slogctx.FromCtx(ctx).Debug("Could not unescape process path string.", slog.Any("error", err))
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
						slogctx.FromCtx(ctx).Warn("Could not create OOM event.", slog.Any("error", err))
					} else {
						eventCh <- entity
					}
				}
			}
		}
	}()

	return eventCh, nil
}

func (w *OOMEventsWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}
