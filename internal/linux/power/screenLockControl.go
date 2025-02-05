// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package power

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/eclipse/paho.golang/paho"
	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	screenLockControlsWorkerPrefID = controlsPrefPrefix + "screen_lock_controls"
)

var (
	ErrInitScreenLockControls = errors.New("could not init screen lock controls worker")
	ErrUnsupportedDesktop     = errors.New("unsupported desktop environment")
)

type screenLockControlsWorker struct {
	prefs *preferences.CommonWorkerPrefs
}

// screenControlCommand represents the D-Bus and MQTT Button Entity information
// for a screen lock control. This information is used to derive the appropriate
// D-Bus call and MQTT button entity config.
type screenControlCommand struct {
	bus    *dbusx.Bus
	name   string
	id     string
	icon   string
	intr   string
	path   string
	method string
}

// execute represents the D-Bus method call to execute the screen control.
func (c *screenControlCommand) execute(ctx context.Context) error {
	err := dbusx.NewMethod(c.bus, c.intr, c.path, c.method).Call(ctx)
	if err != nil {
		return fmt.Errorf("failed to issuse screen control commands %s: %w", c.name, err)
	}

	return nil
}

// setupCommands will generate the appropriate screen control commands based on
// the desktop environment. Some environments can use systemd-logind which
// provides lock and unlock methods while others implement the older
// xscreensaver lock method.
//
//nolint:lll
func setupCommands(_ context.Context, sessionBus *dbusx.Bus, systemBus *dbusx.Bus, device *mqtthass.Device) ([]*screenControlCommand, error) {
	var commands []*screenControlCommand

	desktop := os.Getenv("XDG_CURRENT_DESKTOP")

	switch {
	case strings.Contains(desktop, "KDE"), strings.Contains(desktop, "GNOME"):
		sessionPath, err := systemBus.GetSessionPath()
		if err != nil {
			return nil, fmt.Errorf("unable to set up screen control commands: %w", err)
		}

		// KDE and Gnome can use systemd-logind session lock/unlock on the
		// system bus.
		commands = append(commands,
			&screenControlCommand{
				name:   "Lock Session",
				id:     device.Name + "_lock_session",
				icon:   screenLockedIcon,
				intr:   loginBaseInterface,
				path:   sessionPath,
				method: sessionInterface + ".Lock",
				bus:    systemBus,
			},
			&screenControlCommand{
				name:   "Unlock Session",
				id:     device.Name + "_unlock_session",
				icon:   screenUnlockedIcon,
				intr:   loginBaseInterface,
				path:   sessionPath,
				method: sessionInterface + ".UnLock",
				bus:    systemBus,
			},
		)
	case strings.Contains(desktop, "XFCE"):
		// Xfce implements the screensaver methods on the session bus.
		commands = append(commands,
			&screenControlCommand{
				name:   "Activate Screensaver",
				id:     device.Name + "_activate_screensaver",
				icon:   screenLockedIcon,
				intr:   "org.xfce.ScreenSaver",
				path:   "/",
				method: "org.xfce.ScreenSaver.Lock",
				bus:    sessionBus,
			})
	case strings.Contains(desktop, "Cinnamon"):
		// Cinnamon implements the screensaver methods on the session bus.
		commands = append(commands,
			&screenControlCommand{
				name:   "Activate Screensaver",
				id:     device.Name + "_activate_screensaver",
				icon:   screenLockedIcon,
				intr:   "org.cinnamon.ScreenSaver",
				path:   "/org/cinnamon/ScreenSaver",
				method: "org.cinnamon.ScreenSaver.Lock",
				bus:    sessionBus,
			})
	default:
		return nil, ErrUnsupportedDesktop
	}

	return commands, nil
}

func (w *screenLockControlsWorker) PreferencesID() string {
	return screenLockControlsWorkerPrefID
}

func (w *screenLockControlsWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

// NewScreenLockControl is called by the OS controller of the agent to generate
// MQTT button entities for the screen lock controls.
func NewScreenLockControl(ctx context.Context, device *mqtthass.Device) ([]*mqtthass.ButtonEntity, error) {
	var err error

	worker := &screenLockControlsWorker{}

	worker.prefs, err = preferences.LoadWorker(worker)
	if err != nil {
		return nil, errors.Join(ErrInitScreenLockControls, err)
	}

	if worker.prefs.IsDisabled() {
		return nil, nil
	}

	// Retrieve the D-Bus session bus. Needed by some controls.
	sessionBus, ok := linux.CtxGetSessionBus(ctx)
	if !ok {
		return nil, errors.Join(ErrInitScreenLockControls, linux.ErrNoSessionBus)
	}

	// Retrieve the D-Bus system bus. Needed by some controls.
	systemBus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return nil, errors.Join(ErrInitScreenLockControls, linux.ErrNoSystemBus)
	}

	// Decorate a logger for this controller.
	logger := logging.FromContext(ctx).WithGroup("screensaver_control")

	commands, err := setupCommands(ctx, sessionBus, systemBus, device)
	if err != nil {
		return nil, errors.Join(ErrInitScreenLockControls, err)
	}

	buttons := make([]*mqtthass.ButtonEntity, 0, len(commands))

	for _, command := range commands {
		buttons = append(buttons,
			mqtthass.NewButtonEntity().
				WithDetails(
					mqtthass.App(preferences.AppName),
					mqtthass.Name(command.name),
					mqtthass.ID(command.id),
					mqtthass.DeviceInfo(device),
					mqtthass.Icon(command.icon),
				).
				WithCommand(
					mqtthass.CommandCallback(func(_ *paho.Publish) {
						if err := command.execute(ctx); err != nil {
							logger.Error("Could not execute screen control command.",
								slog.String("name", command.name),
								slog.String("path", command.path),
								slog.String("interface", command.intr),
								slog.String("method", command.method),
								slog.Any("error", err))
						}
					}),
				),
		)
	}

	return buttons, nil
}
