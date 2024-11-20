// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package mem

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/joshuar/go-hass-agent/internal/hass/event"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	oomEventsWorkerID = "oom_events_worker"
	oomDBusPath       = "/org/freedesktop/systemd1/unit"
	unitPathPrefix    = "/org/freedesktop/systemd1/unit"
	oomEventName      = "oom_event"
)

type oomEventData struct {
	Process string `json:"process"`
	PID     int    `json:"pid"`
}

type OOMEventsWorker struct {
	triggerCh chan dbusx.Trigger
	linux.EventWorker
}

//nolint:gocognit
func (w *OOMEventsWorker) Events(ctx context.Context) (<-chan event.Event, error) {
	eventCh := make(chan event.Event)

	go func() {
		defer close(eventCh)

		for {
			select {
			case <-ctx.Done():
				return
			case trigger := <-w.triggerCh:
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
					eventCh <- event.Event{
						EventType: oomEventName,
						EventData: oomEventData{
							Process: processStr,
							PID:     pid,
						},
					}
				}
			}
		}
	}()

	return eventCh, nil
}

func NewOOMEventsWorker(ctx context.Context) (*linux.EventWorker, error) {
	worker := linux.NewEventWorker(oomEventsWorkerID)

	bus, ok := linux.CtxGetSessionBus(ctx)
	if !ok {
		return worker, linux.ErrNoSessionBus
	}

	eventWorker := &OOMEventsWorker{}

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPathNamespace(oomDBusPath),
		dbusx.MatchPropChanged(),
	).Start(ctx, bus)
	if err != nil {
		return nil, fmt.Errorf("unable to set-up D-Bus watch for OOM events: %w", err)
	}

	eventWorker.triggerCh = triggerCh

	worker.EventType = eventWorker

	return worker, nil
}
