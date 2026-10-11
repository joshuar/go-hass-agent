package media

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"

	"github.com/eclipse/paho.golang/paho"
	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/config"
	"github.com/joshuar/go-hass-agent/models"
)

const (
	volIcon  = "mdi:knob"
	muteIcon = "mdi:volume-mute"

	minVolPc  = 0
	maxVolPc  = 100
	volStepPc = 1
)

// runOsascript and runOsascriptOutput are package-level vars so tests can
// substitute them rather than actually changing the volume on the machine
// running the tests.
var (
	runOsascript = func(ctx context.Context, script string) error {
		return exec.CommandContext(ctx, "osascript", "-e", script).Run() //nolint:gosec
	}
	runOsascriptOutput = func(ctx context.Context, script string) (string, error) {
		out, err := exec.CommandContext(ctx, "osascript", "-e", script).Output() //nolint:gosec
		return strings.TrimSpace(string(out)), err
	}
)

// VolumeWorker provides MQTT controls for the volume and mute state of the
// default audio output device.
type VolumeWorker struct {
	*models.WorkerMetadata

	MsgCh         chan mqttapi.Msg
	VolumeControl *mqtthass.NumberEntity[int]
	MuteControl   *mqtthass.SwitchEntity
}

// NewVolumeWorker creates a new MQTT volume/mute controller.
func NewVolumeWorker(ctx context.Context, device *mqtthass.Device) (*VolumeWorker, error) {
	worker := &VolumeWorker{
		WorkerMetadata: models.SetWorkerMetadata("volume_control", "Volume control"),
		MsgCh:          make(chan mqttapi.Msg),
	}

	defaultPrefs := &VolumeWorkerPrefs{}

	prefs, err := workers.LoadWorkerPreferences(volumeControlPreferencesID, defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	if prefs.IsDisabled() {
		return worker, nil
	}

	worker.VolumeControl = mqtthass.NewNumberEntity[int]().
		WithMin(minVolPc).
		WithMax(maxVolPc).
		WithStep(volStepPc).
		WithMode(mqtthass.NumberSlider).
		WithDetails(
			mqtthass.App(config.AppName+"_"+device.Name),
			mqtthass.Name("Volume"),
			mqtthass.ID(device.Name+"_volume"),
			mqtthass.DeviceInfo(device),
			mqtthass.Icon(volIcon),
		).
		WithState(
			mqtthass.StateCallback(func(_ ...any) (json.RawMessage, error) {
				return worker.volumeStateCallback(ctx)
			}),
			mqtthass.ValueTemplate("{{ value_json.value }}"),
		).
		WithCommand(
			mqtthass.CommandCallback(func(p *paho.Publish) {
				worker.volumeCommandCallback(ctx, p)
			}),
		)

	worker.MuteControl = mqtthass.NewSwitchEntity().
		OptimisticMode().
		WithDetails(
			mqtthass.App(config.AppName+"_"+device.Name),
			mqtthass.Name("Mute"),
			mqtthass.ID(device.Name+"_mute"),
			mqtthass.DeviceInfo(device),
			mqtthass.Icon(muteIcon),
		).
		WithState(
			mqtthass.StateCallback(func(_ ...any) (json.RawMessage, error) {
				return worker.muteStateCallback(ctx)
			}),
			mqtthass.ValueTemplate("{{ value }}"),
		).
		WithCommand(
			mqtthass.CommandCallback(func(p *paho.Publish) {
				worker.muteCommandCallback(ctx, p)
			}),
		)

	return worker, nil
}

func (w *VolumeWorker) volumeStateCallback(ctx context.Context) (json.RawMessage, error) {
	out, err := runOsascriptOutput(ctx, "output volume of (get volume settings)")
	if err != nil {
		return json.RawMessage(`{"value": 0}`), fmt.Errorf("could not get volume: %w", err)
	}

	vol, err := strconv.Atoi(out)
	if err != nil {
		return json.RawMessage(`{"value": 0}`), fmt.Errorf("could not parse volume %q: %w", out, err)
	}

	return json.RawMessage(`{"value": ` + strconv.Itoa(vol) + `}`), nil
}

func (w *VolumeWorker) volumeCommandCallback(ctx context.Context, p *paho.Publish) {
	newValue, err := strconv.Atoi(string(p.Payload))
	if err != nil {
		slog.Warn("Could not parse new volume level.", slog.Any("error", err))
		return
	}

	if err := runOsascript(ctx, fmt.Sprintf("set volume output volume %d", newValue)); err != nil {
		slog.Error("Could not set volume level.", slog.Any("error", err))
		return
	}

	w.publish(w.VolumeControl)
}

func (w *VolumeWorker) muteStateCallback(ctx context.Context) (json.RawMessage, error) {
	out, err := runOsascriptOutput(ctx, "output muted of (get volume settings)")
	if err != nil {
		return json.RawMessage(`OFF`), fmt.Errorf("could not get mute state: %w", err)
	}

	if out == "true" {
		return json.RawMessage(`ON`), nil
	}

	return json.RawMessage(`OFF`), nil
}

func (w *VolumeWorker) muteCommandCallback(ctx context.Context, p *paho.Publish) {
	var (
		script string
		err    error
	)

	switch string(p.Payload) {
	case "ON":
		script = "set volume with output muted"
	case "OFF":
		script = "set volume without output muted"
	default:
		return
	}

	if err = runOsascript(ctx, script); err != nil {
		slog.Error("Could not set mute state.", slog.Any("error", err))
		return
	}

	w.publish(w.MuteControl)
}

// stateEntity is satisfied by any entity whose current state can be
// marshalled for publishing.
type stateEntity interface {
	MarshalState(args ...any) (*mqttapi.Msg, error)
}

func (w *VolumeWorker) publish(entity stateEntity) {
	go func() {
		msg, err := entity.MarshalState()
		if err != nil {
			slog.Error("Could not marshal entity state.", slog.Any("error", err))
			return
		}
		w.MsgCh <- *msg
	}()
}
