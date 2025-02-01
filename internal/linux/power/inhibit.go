// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package power

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"syscall"

	"github.com/eclipse/paho.golang/paho"
	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	inhibitWorkerID     = "inhibit_control"
	inhibitWorkerPrefID = controlsPrefPrefix + "inhibit_controls"
)

type inhibitControlWorker struct {
	prefs  *preferences.CommonWorkerPrefs
	entity *mqtthass.SwitchEntity
	fd     int
	logger *slog.Logger
	msgCh  chan *mqttapi.Msg
	bus    *dbusx.Bus
}

func (w *inhibitControlWorker) PreferencesID() string {
	return inhibitWorkerPrefID
}

func (w *inhibitControlWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

//nolint:nilnil
func NewInhibitControl(ctx context.Context, msgCh chan *mqttapi.Msg, device *mqtthass.Device) (*mqtthass.SwitchEntity, error) {
	var err error

	worker := &inhibitControlWorker{
		logger: logging.FromContext(ctx).WithGroup(inhibitWorkerID),
		msgCh:  msgCh,
	}

	// Create an MQTT switch entity for toggling the inhibit lock.
	worker.entity = mqtthass.NewSwitchEntity().
		OptimisticMode().
		WithDetails(
			mqtthass.App(preferences.AppName),
			mqtthass.Name("Inhibit Sleep/Shutdown"),
			mqtthass.ID(device.Name+"_inhibit_lock"),
			mqtthass.DeviceInfo(device),
			mqtthass.Icon("mdi:lock"),
		).
		WithState(
			mqtthass.StateCallback(worker.inhibitStateCallback),
			mqtthass.ValueTemplate("{{ value }}"),
		).
		WithCommand(
			mqtthass.CommandCallback(worker.inhibitCommandCallback),
		)

	worker.prefs, err = preferences.LoadWorker(worker)
	if err != nil {
		return nil, fmt.Errorf("could not load preferences: %w", err)
	}

	if worker.prefs.IsDisabled() {
		return nil, nil
	}

	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return nil, linux.ErrNoSystemBus
	}

	worker.bus = bus

	// On agent shutdown, release any inhibit lock currently held.
	go func() {
		<-ctx.Done()

		if err := worker.releaseInhibitLock(); err != nil {
			worker.logger.Error("Could not release inhibit state.",
				slog.Any("error", err))
		}
	}()

	go func() {
		if err := worker.publishState(worker.msgCh); err != nil {
			worker.logger.Warn("Could not publish initial inhibit state.",
				slog.Any("error", err))
		}
	}()

	return worker.entity, nil
}

// inhibitStateCallback is executed when the inhibit state is read on MQTT.
func (w *inhibitControlWorker) inhibitStateCallback(_ ...any) (json.RawMessage, error) {
	if w.fd > 0 {
		return json.RawMessage(`ON`), nil
	}

	return json.RawMessage(`OFF`), nil
}

// inhibitCommandCallback is executed when the inhibit control is toggled.
func (w *inhibitControlWorker) inhibitCommandCallback(p *paho.Publish) {
	var err error

	state := string(p.Payload)
	switch state {
	case "ON":
		err = w.createInhibitLock()
	case "OFF":
		err = w.releaseInhibitLock()
	}

	if err != nil {
		w.logger.Error("Could not set inhibit state.",
			slog.Any("error", err))

		return
	}

	go func() {
		if err := w.publishState(w.msgCh); err != nil {
			w.logger.Error("Failed to publish mute state to MQTT.", slog.Any("error", err))
		}
	}()
}

// publishState will publish on MQTT the current state of the inhibit lock
// controlled by the worker.
func (w *inhibitControlWorker) publishState(msgCh chan *mqttapi.Msg) error {
	msg, err := w.entity.MarshalState()
	if err != nil {
		return fmt.Errorf("could not marshal entity state: %w", err)
	}
	msgCh <- msg

	return nil
}

// releaseInhibitLock will release any fd file lock the inhibit worker has been
// granted.
func (w *inhibitControlWorker) releaseInhibitLock() error {
	if err := syscall.Close(w.fd); err != nil {
		return fmt.Errorf("error closing inhibit file descriptor lock: %w", err)
	}

	return nil
}

// createInhibitLock will create an inhibit lock for the worker.
func (w *inhibitControlWorker) createInhibitLock() error {
	fd, err := dbusx.GetData[int](w.bus,
		"/org/freedesktop/login1",
		"org.freedesktop.login1",
		"org.freedesktop.login1.Manager.Inhibit",
		"sleep:shutdown",
		preferences.AppName,
		"User requested",
		"block",
	)
	if err != nil {
		return fmt.Errorf("could not create inhibit lock: %w", err)
	}

	w.fd = fd

	return nil
}
