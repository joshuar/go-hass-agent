package power

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"

	"github.com/eclipse/paho.golang/paho"
	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/config"
)

const powerIcon = "mdi:power"

// runOsascript executes an AppleScript one-liner via osascript. It's a
// package-level var so tests can substitute it rather than actually
// sleeping/restarting/shutting down the machine running the tests.
var runOsascript = func(ctx context.Context, script string) error {
	return exec.CommandContext(ctx, "osascript", "-e", script).Run() //nolint:gosec
}

// powerCommand represents a single power control button: its Home
// Assistant display details, and the AppleScript to run when pressed.
type powerCommand struct {
	name   string
	idSfx  string
	icon   string
	script string
}

// powerCommands are the power controls available on macOS. Two of Linux's
// controls have no clean macOS equivalent and are deliberately omitted:
// Hibernate (there's no user-triggerable equivalent -- "safe sleep" is an
// automatic OS decision, not a command) and Unlock Screen (deliberately not
// scriptable by Apple; there's no password/biometric bypass).
var powerCommands = []powerCommand{
	{
		name:   "Lock Screen",
		idSfx:  "lock_screen",
		icon:   "mdi:eye-lock",
		script: `tell application "System Events" to keystroke "q" using {command down, control down}`,
	},
	{
		name:   "Sleep",
		idSfx:  "sleep",
		icon:   "mdi:power-sleep",
		script: `tell application "System Events" to sleep`,
	},
	{
		name:   "Restart",
		idSfx:  "restart",
		icon:   "mdi:restart",
		script: `tell application "System Events" to restart`,
	},
	{
		name:   "Shut Down",
		idSfx:  "shut_down",
		icon:   powerIcon,
		script: `tell application "System Events" to shut down`,
	},
}

// NewPowerControl creates MQTT button entities for the power controls
// available on macOS: locking the screen, sleep, restart and shut down.
func NewPowerControl(ctx context.Context, device *mqtthass.Device) ([]*mqtthass.ButtonEntity, error) {
	defaultPrefs := &Prefs{}

	prefs, err := workers.LoadWorkerPreferences(powerControlPreferencesID, defaultPrefs)
	if err != nil {
		return nil, fmt.Errorf("load preferences: %w", err)
	}

	if prefs.IsDisabled() {
		return nil, nil
	}

	logger := slogctx.FromCtx(ctx).WithGroup("power_control")

	buttons := make([]*mqtthass.ButtonEntity, 0, len(powerCommands))

	for _, command := range powerCommands {
		buttons = append(buttons,
			mqtthass.NewButtonEntity().
				WithDetails(
					mqtthass.App(config.AppName+"_"+device.Name),
					mqtthass.Name(command.name),
					mqtthass.ID(device.Name+"_"+command.idSfx),
					mqtthass.DeviceInfo(device),
					mqtthass.Icon(command.icon),
				).
				WithCommand(
					mqtthass.CommandCallback(func(_ *paho.Publish) {
						pressPowerCommand(ctx, logger, command)
					}),
				),
		)
	}

	return buttons, nil
}

// pressPowerCommand runs a power command's script, logging any error.
func pressPowerCommand(ctx context.Context, logger *slog.Logger, command powerCommand) {
	if err := runOsascript(ctx, command.script); err != nil {
		logger.Error("Could not execute power control.",
			slog.String("control", command.name),
			slog.Any("error", err))
	}
}
