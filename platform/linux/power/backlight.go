// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package power

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eclipse/paho.golang/paho"
	"github.com/godbus/dbus/v5/introspect"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/config"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
	"github.com/joshuar/go-hass-agent/platform/linux"

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"
)

const (
	minBrightnessPc = 0
	maxBrightnessPc = 100
)

var ddcutilPath string

var findDDCUtil = sync.OnceValue(func() error {
	var err error
	ddcutilPath, err = exec.LookPath("ddcutil")
	if err != nil {
		return fmt.Errorf("find ddcutil executable: %w", err)
	}
	return nil
})

type BacklightWorker struct {
	*models.WorkerMetadata

	bus   *dbusx.Bus
	prefs *workers.CommonWorkerPrefs

	Entity *mqtthass.NumberEntity[int]
	MsgCh  chan mqttapi.Msg

	desktop string
}

// NewBacklightControl creates an entity worker that can manipulate the screen backlight brightness.
func NewBacklightControl(ctx context.Context, device *mqtthass.Device) (*BacklightWorker, error) {
	desktop := os.Getenv("XDG_CURRENT_DESKTOP")
	switch {
	case strings.Contains(desktop, "GNOME"), strings.Contains(desktop, "KDE"):
	default:
		if err := findDDCUtil(); err != nil {
			return nil, fmt.Errorf("new backlight control: %w", linux.ErrUnsupportedDesktop)
		}
	}

	worker := &BacklightWorker{
		WorkerMetadata: models.SetWorkerMetadata("backlight", "Backlight (screen brightness)"),
		MsgCh:          make(chan mqttapi.Msg),
		desktop:        desktop,
	}

	var ok bool
	worker.bus, ok = linux.CtxGetSessionBus(ctx)
	if !ok {
		return nil, fmt.Errorf("get session bus: %w", linux.ErrNoSystemBus)
	}

	if !hasBrightnessControls(worker.desktop, worker.bus) {
		return nil, fmt.Errorf("check for brightness controls: %w", linux.ErrUnsupportedDesktop)
	}

	defaultPrefs := &workers.CommonWorkerPrefs{}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(sensorsPrefPrefix+"screen_lock", defaultPrefs)
	if err != nil {
		return nil, fmt.Errorf("load preferences: %w", err)
	}

	// Generate a number entity for the brightness control.
	worker.Entity = mqtthass.NewNumberEntity[int]().
		WithMin(minBrightnessPc).
		WithMax(maxBrightnessPc).
		WithStep(1).
		WithMode(mqtthass.NumberSlider).
		WithDetails(
			mqtthass.App(config.AppName+"_"+device.Name),
			mqtthass.Name("Screen brightness"),
			mqtthass.ID(device.Name+"_backlight"),
			mqtthass.DeviceInfo(device),
			mqtthass.Icon("mdi:brightness-percent"),
		).
		WithState(
			mqtthass.StateCallback(func(_ ...any) (json.RawMessage, error) {
				return worker.stateCallback(ctx)
			}),
			mqtthass.ValueTemplate("{{ value_json.value }}"),
			mqtthass.Units("%"),
		).
		WithCommand(
			mqtthass.CommandCallback(func(p *paho.Publish) {
				brightness, err := strconv.Atoi(string(p.Payload))
				if err != nil {
					slogctx.FromCtx(ctx).Warn("Could not parse screen brightness level.",
						slog.Any("error", err),
					)
					return
				}
				slogctx.FromCtx(ctx).Debug("Adjusting screen brightness.",
					slog.Int("brightness", brightness),
				)
				err = worker.controlCallback(ctx, brightness)
				if err != nil {
					slogctx.FromCtx(ctx).Warn("Could not adjust screen brightness.",
						slog.Any("error", err))
				}
			}),
		)

	// Publish the current display brightness.
	go worker.publishState(ctx)

	// Monitor for external changes.
	go worker.monitor(ctx)

	return worker, nil
}

func (w *BacklightWorker) publishState(ctx context.Context) {
	msg, err := w.Entity.MarshalState()
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Unable to publish backlight state",
			slog.Any("error", err),
		)
	}
	w.MsgCh <- *msg
}

// stateCallback is called when an MQTT message is published to get the current brightness.
func (w *BacklightWorker) stateCallback(ctx context.Context) (json.RawMessage, error) {
	var (
		brightness int
		err        error
	)
	switch {
	case strings.Contains(w.desktop, "GNOME"):
		brightness, err = getBrightnessGnome(w.bus)
	case strings.Contains(w.desktop, "KDE"):
		brightness, err = getBrightnessKDE(w.bus)
	default:
		brightness, err = getBrightnessDDCUtil(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("get brightness: %w", err)
	}

	return json.RawMessage(`{ "value": ` + strconv.Itoa(brightness) + ` }`), nil
}

// controlCallback is called when an MQTT message is published to change the brightness.
func (w *BacklightWorker) controlCallback(ctx context.Context, value int) error {
	var err error
	switch {
	case strings.Contains(w.desktop, "GNOME"):
		err = setBrightnessGnome(ctx, w.bus, value)
	case strings.Contains(w.desktop, "KDE"):
		err = setBrightnessKDE(ctx, w.bus, value)
	default:
		err = setBrightnessDDCUtil(ctx, value)
	}

	if err != nil {
		return fmt.Errorf("set brightness: %w", err)
	}

	return nil
}

// monitor will set up a way to monitor for external brightness changes.
func (w *BacklightWorker) monitor(ctx context.Context) {
	var err error
	switch {
	case strings.Contains(w.desktop, "GNOME"):
		err = w.monitorBrightnessGnome(ctx)
	case strings.Contains(w.desktop, "KDE"):
		err = w.monitorBrightnessKDE(ctx)
	default:
		w.monitorBrightnessDDCUtil(ctx)
	}
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Unable to monitor brightness.",
			slog.Any("error", err),
		)
	}
}

func hasBrightnessControls(desktop string, bus *dbusx.Bus) bool {
	switch {
	case strings.Contains(desktop, "GNOME"):
		hasBrightness, err := dbusx.NewProperty[bool](bus,
			"/org/gnome/Shell/Brightness",
			"org.gnome.Shell.Brightness",
			"org.gnome.Shell.Brightness.HasBrightnessControl",
		).Get()
		if err != nil {
			return false
		}
		return hasBrightness
	case strings.Contains(desktop, "KDE"):
		introspection, err := dbusx.NewIntrospection(
			bus,
			"org.kde.Solid.PowerManagement",
			"/org/kde/Solid/PowerManagement/Actions/BrightnessControl",
		)
		if err != nil {
			return false
		}
		if slices.ContainsFunc(introspection.Interfaces, func(i introspect.Interface) bool {
			if i.Name == "org.kde.Solid.PowerManagement.Actions.BrightnessControl" {
				return slices.ContainsFunc(i.Methods, func(m introspect.Method) bool {
					return m.Name == "setBrightness"
				})
			}
			return false
		}) {
			return true
		}
	default:
		if err := findDDCUtil(); err != nil {
			return false
		}
	}
	return false
}

// setBrightnessGnome will use D-Bus to change the brightness on Gnome desktops.
func setBrightnessGnome(ctx context.Context, bus *dbusx.Bus, value int) error {
	if err := dbusx.NewMethod(bus,
		"/org/gnome/Shell/Brightness",
		"org.gnome.Shell.Brightness",
		"org.freedesktop.DBus.Properties.Set").
		Call(ctx, "org.gnome.SettingsDaemon.Power.Screen", "Brightness", value); err != nil {
		return fmt.Errorf("set brightness from Gnome: %w", err)
	}
	return nil
}

// getBrightnessKDE will fetch the current brightness using KDE D-Bus methods.
func getBrightnessGnome(bus *dbusx.Bus) (int, error) {
	brightness, err := dbusx.GetData[int](bus,
		"/org/gnome/Shell/Brightness",
		"org.gnome.Shell.Brightness",
		"org.freedesktop.DBus.Properties.Get",
		"org.gnome.Shell.Brightness", "Brightness",
	)
	if err != nil {
		return 0, fmt.Errorf("get brightness from Gnome: %w", err)
	}
	return brightness, nil
}

// monitorBrightnessGnome sets up a D-Bus watch for changes to brightness on the Gnome desktop.
func (w *BacklightWorker) monitorBrightnessGnome(ctx context.Context) error {
	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath("/org/gnome/Shell/Brightness"),
		dbusx.MatchPropChanged(),
	).Start(ctx, w.bus)
	if err != nil {
		return fmt.Errorf("watch brightness Gnome: %w", err)
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-triggerCh:
				props, err := dbusx.ParsePropertiesChanged(event.Content)
				if err != nil {
					slogctx.FromCtx(ctx).Warn("Do not understand changed property.",
						slog.Any("error", err),
					)
					continue
				}
				if _, found := props.Changed["Brightness"]; found {
					w.publishState(ctx)
				}
			}
		}
	}()
	slogctx.FromCtx(ctx).Debug("Monitoring screen brightness via KDE.")
	return nil
}

// setBrightnessKDE will use D-Bus to change the brightness on KDE desktops.
func setBrightnessKDE(ctx context.Context, bus *dbusx.Bus, value int) error {
	if err := dbusx.NewMethod(bus,
		"org.kde.Solid.PowerManagement",
		"/org/kde/Solid/PowerManagement/Actions/BrightnessControl",
		"org.kde.Solid.PowerManagement.Actions.BrightnessControl.setBrightness").
		Call(ctx, value*maxBrightnessPc); err != nil {
		return fmt.Errorf("set brightness from KDE: %w", err)
	}
	return nil
}

// getBrightnessKDE will fetch the current brightness using KDE D-Bus methods.
func getBrightnessKDE(bus *dbusx.Bus) (int, error) {
	brightness, err := dbusx.GetData[int](bus,
		"/org/kde/Solid/PowerManagement/Actions/BrightnessControl",
		"org.kde.Solid.PowerManagement",
		"org.kde.Solid.PowerManagement.Actions.BrightnessControl.brightness")
	if err != nil {
		return 0, fmt.Errorf("get brightness from KDE: %w", err)
	}
	return brightness / maxBrightnessPc, nil
}

// monitorBrightnessKDE sets up a D-Bus watch for changes to brightness on the KDE desktop.
func (w *BacklightWorker) monitorBrightnessKDE(ctx context.Context) error {
	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath("/org/kde/Solid/PowerManagement/Actions/BrightnessControl"),
		dbusx.MatchInterface("org.kde.Solid.PowerManagement.Actions.BrightnessControl"),
		dbusx.MatchMembers("brightnessChanged"),
	).Start(ctx, w.bus)
	if err != nil {
		return fmt.Errorf("watch brightness KDE: %w", err)
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-triggerCh:
				w.publishState(ctx)
			}
		}
	}()
	slogctx.FromCtx(ctx).Debug("Monitoring screen brightness via KDE.")
	return nil
}

// setBrightnessDDCUtil will use the ddcutil program to set the brightness.
func setBrightnessDDCUtil(ctx context.Context, value int) error {
	_, err := exec.CommandContext(ctx, ddcutilPath, "-t", "setvcp", "10", strconv.Itoa(value)).Output()
	if err != nil {
		return fmt.Errorf("ddcutil setvcp 10: %w", err)
	}
	return nil
}

// getBrightnessDDCUtil will use the ddcutil program to get the brightness.
func getBrightnessDDCUtil(ctx context.Context) (int, error) {
	output, err := exec.CommandContext(ctx, ddcutilPath, "-t", "getvcp", "10").Output()
	if err != nil {
		return 0, fmt.Errorf("ddcutil getvcp 10: %w", err)
	}

	values := strings.Split(string(output), " ")
	if len(values) < 3 {
		return 0, errors.New("invalid ddcutil output")
	}
	b, err := strconv.Atoi(values[3])
	if err != nil {
		return 0, fmt.Errorf("ddcutil getvcp 10: %w", err)
	}

	return b, nil
}

// monitorBrightnessDDCUtil will monitor brightness changes by polling with ddcutil on an interval. Not very efficient
// and not real time, but works on any system with ddcutil installed...
func (w *BacklightWorker) monitorBrightnessDDCUtil(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				w.publishState(ctx)
			}
		}
	}()

	slogctx.FromCtx(ctx).Debug("Monitoring screen brightness by polling with ddcutil, every minute.")
}
