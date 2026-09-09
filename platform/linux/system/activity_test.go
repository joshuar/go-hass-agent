// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package system

import (
	"bufio"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

// stateWaitTimeout is how long a test will wait for the worker to produce a
// sensor update.
const (
	stateWaitTimeout = 5 * time.Second

	// inputRetryInterval is how often a test generates another input event
	// while waiting for the worker to report.
	inputRetryInterval = 50 * time.Millisecond

	// shortIdleTimeout is used where a test wants the idle timer to fire.
	shortIdleTimeout = 100 * time.Millisecond
)

// newTestActivityWorker builds an activityWorker with no input devices,
// bypassing NewUserActivitySensor so the test needs neither evdev access nor
// the input group. The idle timeout is deliberately long: if the sensor goes
// off during a test, it is because of a sleep/shutdown signal and never
// because of the idle timer.
func newTestActivityWorker(bus *dbusx.Bus) *activityWorker {
	return &activityWorker{
		WorkerMetadata: models.SetWorkerMetadata(activityWorkerID, "User Activity"),
		prefs:          &activityWorkerPrefs{IdleTimeout: time.Hour.String()},
		bus:            bus,
		activity:       make(chan bool, 1),
	}
}

// startTestBus runs a private D-Bus daemon and points the D-Bus client at it,
// so tests never touch the real system bus. The test is skipped where no
// daemon is available, such as a CI runner without D-Bus installed.
func startTestBus(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("dbus-daemon"); err != nil {
		t.Skip("dbus-daemon not available, skipping D-Bus based test")
	}

	daemon := exec.CommandContext(t.Context(), "dbus-daemon", "--session", "--nofork", "--print-address")

	stdout, err := daemon.StdoutPipe()
	require.NoError(t, err)
	require.NoError(t, daemon.Start())

	// The daemon is killed when the test context is canceled, so it only
	// needs reaping here.
	t.Cleanup(func() {
		if waitErr := daemon.Wait(); waitErr != nil {
			t.Logf("test bus stopped: %v", waitErr)
		}
	})

	address, err := bufio.NewReader(stdout).ReadString('\n')
	require.NoError(t, err)

	t.Setenv("DBUS_SYSTEM_BUS_ADDRESS", strings.TrimSpace(address))
}

// startTestWorker starts a worker attached to a private bus, and returns it
// with the channel it sends sensors on.
func startTestWorker(t *testing.T) (*activityWorker, <-chan models.Entity) {
	t.Helper()

	startTestBus(t)

	bus, err := dbusx.NewBus(t.Context(), dbusx.SystemBus)
	require.NoError(t, err)

	worker := newTestActivityWorker(bus)

	sensorCh, err := worker.Start(t.Context())
	require.NoError(t, err)

	return worker, sensorCh
}

// emitSignal broadcasts a logind-style signal on the private bus.
func emitSignal(t *testing.T, member string, value bool) {
	t.Helper()

	conn, err := dbus.ConnectSystemBus()
	require.NoError(t, err)

	defer conn.Close()

	require.NoError(t, conn.Emit(dbus.ObjectPath(loginBasePath), managerInterface+"."+member, value))
}

// requireState waits for the next sensor update and asserts its state.
func requireState(t *testing.T, sensorCh <-chan models.Entity, want bool) {
	t.Helper()

	select {
	case entity, open := <-sensorCh:
		require.True(t, open, "sensor channel closed while waiting for state %v", want)

		sensor, err := entity.AsSensor()
		require.NoError(t, err)
		assert.Equal(t, want, sensor.State)
	case <-time.After(stateWaitTimeout):
		assert.Fail(t, "timed out waiting for sensor state", "wanted state %v", want)
	}
}

// requireNoState asserts the worker stays silent for the given duration.
func requireNoState(t *testing.T, sensorCh <-chan models.Entity, wait time.Duration) {
	t.Helper()

	select {
	case entity := <-sensorCh:
		sensor, err := entity.AsSensor()
		require.NoError(t, err)
		assert.Fail(t, "unexpected sensor update", "state %v", sensor.State)
	case <-time.After(wait):
	}
}

// requireStateAfterInput keeps generating input events until the worker
// reports the wanted state, as a user waking the machine would.
func requireStateAfterInput(t *testing.T, worker *activityWorker, sensorCh <-chan models.Entity, want bool) {
	t.Helper()

	for range int(stateWaitTimeout / inputRetryInterval) {
		select {
		case worker.activity <- true:
		default:
		}

		select {
		case entity := <-sensorCh:
			sensor, err := entity.AsSensor()
			require.NoError(t, err)
			assert.Equal(t, want, sensor.State)

			return
		case <-time.After(inputRetryInterval):
		}
	}

	assert.Fail(t, "timed out waiting for state after input", "wanted state %v", want)
}

func Test_activityWorker_resetOnSleepOrShutdown(t *testing.T) {
	tests := []struct {
		name   string
		signal string
	}{
		{name: "shutdown", signal: shutdownSignal},
		{name: "sleep", signal: sleepSignal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			worker, sensorCh := startTestWorker(t)

			// An input event switches the sensor on.
			worker.activity <- true
			requireState(t, sensorCh, true)

			// The signal announcing the system is going down switches it off
			// again, while the agent can still send it.
			emitSignal(t, tt.signal, true)
			requireState(t, sensorCh, false)
		})
	}
}

func Test_activityWorker_queuedInputDoesNotUndoReset(t *testing.T) {
	tests := []struct {
		name   string
		signal string
	}{
		{name: "shutdown", signal: shutdownSignal},
		{name: "sleep", signal: sleepSignal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			worker, sensorCh := startTestWorker(t)

			worker.activity <- true
			requireState(t, sensorCh, true)

			emitSignal(t, tt.signal, true)
			requireState(t, sensorCh, false)

			// Input devices keep generating events for the moment it takes the
			// system to actually go down. Acting on one of those would switch
			// the sensor back on and leave Home Assistant holding it, which is
			// the whole problem being fixed.
			worker.activity <- true
			requireNoState(t, sensorCh, time.Second)
		})
	}
}

func Test_activityWorker_resumeRestoresReporting(t *testing.T) {
	worker, sensorCh := startTestWorker(t)

	worker.activity <- true
	requireState(t, sensorCh, true)

	emitSignal(t, sleepSignal, true)
	requireState(t, sensorCh, false)

	// On resume the same signal is sent with false, after which activity is
	// reported as normal again. The wake event may be processed before that
	// signal arrives, so input is generated until the sensor reports.
	emitSignal(t, sleepSignal, false)
	requireStateAfterInput(t, worker, sensorCh, true)
}

func Test_activityWorker_worksWithoutSystemBus(t *testing.T) {
	// The agent also runs in containers, where the system bus is often not
	// available. Activity then cannot be reset before the system goes down,
	// but it must still be reported as it was before.
	worker := newTestActivityWorker(nil)
	worker.prefs = &activityWorkerPrefs{IdleTimeout: shortIdleTimeout.String()}

	sensorCh, err := worker.Start(t.Context())
	require.NoError(t, err)

	worker.activity <- true
	requireState(t, sensorCh, true)

	// The idle timer still switches the sensor off.
	requireState(t, sensorCh, false)
}

func Test_activityWorker_ignoresResumeSignal(t *testing.T) {
	worker, sensorCh := startTestWorker(t)

	worker.activity <- true
	requireState(t, sensorCh, true)

	// Neither signal switches the sensor off when it carries false.
	emitSignal(t, sleepSignal, false)
	emitSignal(t, shutdownSignal, false)
	requireNoState(t, sensorCh, time.Second)
}
