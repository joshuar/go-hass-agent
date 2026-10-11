package power

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"sync"

	"github.com/eclipse/paho.golang/paho"
	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/config"
	"github.com/joshuar/go-hass-agent/models"
)

// startCaffeinate and stopCaffeinate are package-level vars so tests can
// substitute them instead of actually preventing sleep on the machine
// running the tests.
var (
	startCaffeinate = func() (*exec.Cmd, error) {
		// -d: prevent display sleep, -i: prevent idle sleep, -s: prevent
		// system sleep while on AC power. No duration, so it holds until
		// the process is killed.
		cmd := exec.Command("caffeinate", "-d", "-i", "-s") //nolint:gosec
		if err := cmd.Start(); err != nil {
			return nil, err
		}

		return cmd, nil
	}
	stopCaffeinate = func(cmd *exec.Cmd) error {
		if cmd == nil || cmd.Process == nil {
			return nil
		}

		return cmd.Process.Kill()
	}
)

// InhibitWorker provides an MQTT switch that prevents the system from
// sleeping while it's on, via the caffeinate command-line tool.
type InhibitWorker struct {
	*models.WorkerMetadata

	InhibitControl *mqtthass.SwitchEntity
	MsgCh          chan mqttapi.Msg

	mu  sync.Mutex
	cmd *exec.Cmd
}

// NewInhibitWorker creates a new MQTT switch for inhibiting sleep.
func NewInhibitWorker(ctx context.Context, device *mqtthass.Device) (*InhibitWorker, error) {
	worker := &InhibitWorker{
		WorkerMetadata: models.SetWorkerMetadata("inhibit_control", "Inhibit sleep control"),
		MsgCh:          make(chan mqttapi.Msg),
	}

	worker.InhibitControl = mqtthass.NewSwitchEntity().
		OptimisticMode().
		WithDetails(
			mqtthass.App(config.AppName+"_"+device.Name),
			mqtthass.Name("Inhibit Sleep"),
			mqtthass.ID(device.Name+"_inhibit_sleep"),
			mqtthass.DeviceInfo(device),
			mqtthass.Icon("mdi:coffee"),
		).
		WithState(
			mqtthass.StateCallback(worker.stateCallback),
			mqtthass.ValueTemplate("{{ value }}"),
		).
		WithCommand(
			mqtthass.CommandCallback(worker.commandCallback),
		)

	defaultPrefs := &Prefs{}

	prefs, err := workers.LoadWorkerPreferences(inhibitControlPreferencesID, defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	if prefs.IsDisabled() {
		return worker, nil
	}

	// Release any held inhibit lock on agent shutdown.
	go func() {
		<-ctx.Done()

		if err := worker.release(); err != nil {
			slogctx.FromCtx(ctx).Error("Could not release inhibit lock.",
				slog.Any("error", err))
		}
	}()

	return worker, nil
}

func (w *InhibitWorker) stateCallback(_ ...any) (json.RawMessage, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.cmd != nil {
		return json.RawMessage(`ON`), nil
	}

	return json.RawMessage(`OFF`), nil
}

func (w *InhibitWorker) commandCallback(p *paho.Publish) {
	var err error

	switch string(p.Payload) {
	case "ON":
		err = w.hold()
	case "OFF":
		err = w.release()
	}

	if err != nil {
		slog.Error("Could not set inhibit state.", slog.Any("error", err))
		return
	}

	go func() {
		msg, err := w.InhibitControl.MarshalState()
		if err != nil {
			slog.Error("Could not marshal inhibit state.", slog.Any("error", err))
			return
		}
		w.MsgCh <- *msg
	}()
}

// hold starts caffeinate, if not already running.
func (w *InhibitWorker) hold() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.cmd != nil {
		return nil
	}

	cmd, err := startCaffeinate()
	if err != nil {
		return fmt.Errorf("could not start caffeinate: %w", err)
	}

	w.cmd = cmd

	return nil
}

// release stops caffeinate, if running.
func (w *InhibitWorker) release() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.cmd == nil {
		return nil
	}

	if err := stopCaffeinate(w.cmd); err != nil {
		return fmt.Errorf("could not stop caffeinate: %w", err)
	}

	w.cmd = nil

	return nil
}
