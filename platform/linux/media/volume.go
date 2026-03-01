// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package media

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"slices"
	"strconv"

	"github.com/eclipse/paho.golang/paho"
	slogctx "github.com/veqryn/slog-context"

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/config"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/pkg/linux/pipewire"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

const (
	muteIcon = "mdi:volume-mute"
	volIcon  = "mdi:knob"

	minVolpc  = 0
	maxVolpc  = 100
	volStepPc = 1

	audioControlPreferencesID = "sensors.media.audio"
)

// VolumeWorker is a struct containing the data for providing audio state
// tracking and control.
type VolumeWorker struct {
	*workers.CommonWorkerPrefs
	*models.WorkerMetadata

	MsgCh         chan mqttapi.Msg
	pwEventChan   chan pipewire.Event
	MuteControl   *mqtthass.SwitchEntity
	VolumeControl *mqtthass.NumberEntity[int]
	// nodes maps PipeWire object ID → NodeState for every Audio/Sink we know.
	nodes map[int]*audioNodeState

	// nodesByName maps node.name → object ID for fast lookup when resolving the
	// default sink name to an actual node.
	nodesByName map[string]int

	// defaultSinkName is the node.name of the current default sink as reported
	// by Metadata.  Empty string means "not yet known".
	defaultSinkName string

	// defaultSinkID is the resolved object ID of the default sink.
	// -1 means "not yet resolved".
	defaultSinkID int
}

// audioNodeState holds what we know about an Audio/Sink node.
type audioNodeState struct {
	Name   string  // node.name – used for matching against metadata
	Desc   string  // human-readable label for display
	Volume float64 // last seen linear volume (-1 = not yet seen)
	Muted  bool
}

// metaSinkValue is the JSON object form of a default.audio.sink metadata value.
type metaSinkValue struct {
	Name string `json:"name"`
}

// entity is a convienience interface to treat all entities the same.
type entity interface {
	MarshalState(args ...any) (*mqttapi.Msg, error)
}

func NewVolumeWorker(ctx context.Context, device *mqtthass.Device) (*VolumeWorker, error) {
	worker := &VolumeWorker{
		WorkerMetadata: models.SetWorkerMetadata("volume_control", "Volume control"),
		MsgCh:          make(chan mqttapi.Msg),
		nodes:          make(map[int]*audioNodeState),
		nodesByName:    make(map[string]int),
	}

	var err error

	defaultPrefs := &workers.CommonWorkerPrefs{}
	worker.CommonWorkerPrefs, err = workers.LoadWorkerPreferences(audioControlPreferencesID, defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	if worker.IsDisabled() {
		return worker, nil
	}

	// Set up pipewire listener.
	monitor, found := linux.CtxGetPipewireMonitor(ctx)
	if !found {
		return worker, errors.New("no pipewire monitor in context")
	}
	worker.pwEventChan = monitor.AddListener(volumePipewireEventFilter)

	id, name, err := pipewire.FindDefaultAudioSink()
	if err != nil {
		return worker, fmt.Errorf("find default audio sink: %w", err)
	}
	worker.defaultSinkID = id
	worker.defaultSinkName = name

	worker.VolumeControl = mqtthass.NewNumberEntity[int]().
		WithMin(minVolpc).
		WithMax(maxVolpc).
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
			mqtthass.StateCallback(worker.volStateCallback),
			mqtthass.ValueTemplate("{{ value_json.value }}"),
		).
		WithCommand(
			mqtthass.CommandCallback(worker.volCommandCallback),
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
			mqtthass.StateCallback(worker.muteStateCallback),
			mqtthass.ValueTemplate("{{ value }}"),
		).
		WithCommand(
			mqtthass.CommandCallback(worker.muteCommandCallback),
		)

	// update := func() { // Pulseaudio changed state. Get the new state.
	// 	// Publish and update mute state if it changed.
	// 	if err := publishAudioState(worker.MsgCh, worker.MuteControl); err != nil {
	// 		slogctx.FromCtx(ctx).Error("Failed to publish mute state to MQTT.", slog.Any("error", err))
	// 	}

	// 	// Publish and update volume if it changed.
	// 	if err := publishAudioState(worker.MsgCh, worker.VolumeControl); err != nil {
	// 		slogctx.FromCtx(ctx).Error("Failed to publish mute state to MQTT.", slog.Any("error", err))
	// 	}
	// }

	// Process Pipewire state updates as they are received.
	go func() {
		for event := range worker.pwEventChan {
			switch event.Type {
			case pipewire.InterfaceNodeEvent:
				worker.handleNode(ctx, event)
			case pipewire.InterfaceMetadataEvent:
				worker.handleMetadata(ctx, event)
			}
		}
	}()

	return worker, nil
}

// volStateCallback is executed when the volume is read on MQTT.
func (d *VolumeWorker) volStateCallback(_ ...any) (json.RawMessage, error) {
	vol, err := pipewire.GetVolume()
	if err != nil {
		return json.RawMessage(`{ "value": 0 }`), err
	}

	slog.Debug("Publishing volume change.", slog.Int("volume", int(vol)))

	return json.RawMessage(`{ "value": ` + strconv.FormatFloat(vol, 'f', 0, 64) + ` }`), nil
}

// volCommandCallback is called when the volume is changed on MQTT.
func (d *VolumeWorker) volCommandCallback(p *paho.Publish) {
	if newValue, err := strconv.Atoi(string(p.Payload)); err != nil {
		slog.Debug("Could not parse new volume level.", slog.Any("error", err))
	} else {
		slog.Debug("Received volume change from Home Assistant.", slog.Int("volume", newValue))

		if err := pipewire.SetVolume(float64(newValue)); err != nil {
			slog.Error("Could not set volume level.", slog.Any("error", err))
			return
		}

		go func() {
			if err := publishAudioState(d.MsgCh, d.VolumeControl); err != nil {
				slog.Error("Failed to publish mute state to MQTT.", slog.Any("error", err))
			}
		}()
	}
}

// muteStateCallback is executed when the mute state is read on MQTT.
func (d *VolumeWorker) muteStateCallback(_ ...any) (json.RawMessage, error) {
	muteVal, err := pipewire.IsMuted()
	if err != nil {
		return json.RawMessage(`OFF`), err
	}

	switch muteVal {
	case true:
		return json.RawMessage(`ON`), nil
	default:
		return json.RawMessage(`OFF`), nil
	}
}

// muteCommandCallback is executed when the mute state is changed on MQTT.
func (d *VolumeWorker) muteCommandCallback(p *paho.Publish) {
	var err error

	switch string(p.Payload) {
	case "ON":
		err = pipewire.Mute()
	case "OFF":
		err = pipewire.Unmute()
	}

	if err != nil {
		slog.Error("Could not set mute state.", slog.Any("error", err))

		return
	}

	go func() {
		if err := publishAudioState(d.MsgCh, d.MuteControl); err != nil {
			slog.Error("Failed to publish mute state to MQTT.", slog.Any("error", err))
		}
	}()
}

func (d *VolumeWorker) handleMetadata(ctx context.Context, e pipewire.Event) {
	for _, m := range e.Metadata {
		if m.Key != "default.audio.sink" {
			continue
		}
		if string(m.Value) == "null" || len(m.Value) == 0 {
			d.defaultSinkName = ""
			d.defaultSinkID = -1
			slogctx.FromCtx(ctx).Debug("Default audio sink cleared.")
			continue
		}
		// PipeWire may encode the value as {"name":"alsa_output...."} or as a
		// plain JSON string.  Try the object form first.
		var obj metaSinkValue
		if err := json.Unmarshal(m.Value, &obj); err == nil && obj.Name != "" {
			d.defaultSinkName = obj.Name
		} else {
			var s string
			if err := json.Unmarshal(m.Value, &s); err == nil {
				d.defaultSinkName = s
			}
		}
		d.resolveDefault(ctx)
	}
}

func (d *VolumeWorker) resolveDefault(ctx context.Context) {
	if d.defaultSinkName == "" {
		return
	}
	id, ok := d.nodesByName[d.defaultSinkName]
	if !ok {
		return // node not yet seen; will be resolved when the node arrives
	}
	if id == d.defaultSinkID {
		return // nothing changed
	}
	d.defaultSinkID = id
	slogctx.FromCtx(ctx).Debug("Default audio sink changed.")
}

func (d *VolumeWorker) handleNode(ctx context.Context, e pipewire.Event) {
	if e.Change == "removed" {
		st, ok := d.nodes[e.ID]
		if !ok {
			return
		}
		if e.ID == d.defaultSinkID {
			slogctx.FromCtx(ctx).Debug("Default sink removed.",
				slog.String("sink_name", audioSinkDisplayName(st)))
			d.defaultSinkID = -1
		}
		delete(d.nodesByName, st.Name)
		delete(d.nodes, e.ID)
		return
	}

	mc := e.Info.Props.MediaClass

	st, known := d.nodes[e.ID]
	if !known {
		// On "changed" diffs MediaClass is typically absent; skip unknown nodes
		// unless we can confirm they are Audio/Sink from this event.
		if mc != "Audio/Sink" {
			return
		}
		st = &audioNodeState{Volume: -1}
		d.nodes[e.ID] = st
	}

	// Update identifying fields when present in this (possibly sparse) diff.
	if e.Info.Props.NodeName != "" && e.Info.Props.NodeName != st.Name {
		if st.Name != "" {
			delete(d.nodesByName, st.Name)
		}
		st.Name = e.Info.Props.NodeName
		d.nodesByName[st.Name] = e.ID
	}
	if e.Info.Props.NodeDesc != "" {
		st.Desc = e.Info.Props.NodeDesc
	} else if e.Info.Props.NodeDesc != "" && st.Desc == "" {
		st.Desc = e.Info.Props.NodeDesc
	}

	// A newly registered node may be the default sink we were waiting for.
	d.resolveDefault(ctx)

	// ── volume / mute ────────────────────────────────────────────────────────
	// Ignore everything that isn't the default sink.
	if e.ID != d.defaultSinkID {
		return
	}

	for _, pp := range e.Info.Params.Props {
		vol, hasVol := avgVolume(pp.Volumes, pp.Volume)

		if hasVol {
			pct := linearToPercent(vol)
			volChanged := math.Abs(vol-st.Volume) > 0.0001

			if volChanged {
				slogctx.FromCtx(ctx).Debug("Volume changed.",
					slog.String("device", audioSinkDisplayName(st)),
					slog.Float64("volume", pct),
				)
				st.Volume = vol
			}

			if pp.Mute != nil && *pp.Mute != st.Muted {
				if *pp.Mute {
					slogctx.FromCtx(ctx).Debug("Muted.",
						slog.String("device", audioSinkDisplayName(st)),
					)
				} else {
					slogctx.FromCtx(ctx).Debug("Unmuted.",
						slog.String("device", audioSinkDisplayName(st)),
					)
				}
				st.Muted = *pp.Mute
			}
		} else if pp.Mute != nil && *pp.Mute != st.Muted {
			// Mute-only update (no volume field in this diff).
			if *pp.Mute {
				slogctx.FromCtx(ctx).Debug("Muted.",
					slog.String("device", audioSinkDisplayName(st)),
				)
			} else {
				slogctx.FromCtx(ctx).Debug("Unmuted.",
					slog.String("device", audioSinkDisplayName(st)),
				)
			}
			st.Muted = *pp.Mute
		}
	}
}

func publishAudioState(msgCh chan mqttapi.Msg, entity entity) error {
	msg, err := entity.MarshalState()
	if err != nil {
		return fmt.Errorf("could not marshal entity state: %w", err)
	}
	msgCh <- *msg

	return nil
}

// displayName returns the most human-readable label for a node.
func audioSinkDisplayName(st *audioNodeState) string {
	if st.Desc != "" {
		return st.Desc
	}
	return st.Name
}

// volumePipewireEventFilter filters the pipewire events.
func volumePipewireEventFilter(event *pipewire.Event) bool {
	switch event.Type {
	case pipewire.InterfaceMetadataEvent:
		if event.Change == "removed" {
			return false
		}
		return slices.ContainsFunc(event.Metadata, func(m pipewire.EventMetadata) bool {
			return m.Key == "default.audio.sink"
		})
	case pipewire.InterfaceNodeEvent:
		return event.Info != nil
	default:
		return false
	}
}

// linearToPercent converts a PipeWire linear amplitude ratio to the
// perceptual percentage shown by GUI mixers (cubic-root scale).
func linearToPercent(linear float64) float64 {
	return math.Cbrt(linear) * 100
}

// avgVolume returns the mean of a per-channel volume slice, falling back to a
// scalar volume value if no channel data is present.
func avgVolume(channels []float64, scalar *float64) (float64, bool) {
	if len(channels) > 0 {
		sum := 0.0
		for _, v := range channels {
			sum += v
		}
		return sum / float64(len(channels)), true
	}
	if scalar != nil {
		return *scalar, true
	}
	return 0, false
}
