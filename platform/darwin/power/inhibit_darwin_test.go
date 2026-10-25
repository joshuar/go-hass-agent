package power

import (
	"context"
	"os/exec"
	"testing"

	"github.com/eclipse/paho.golang/paho"
	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInhibitWorkerToggle(t *testing.T) {
	var started, stopped int

	origStart, origStop := startCaffeinate, stopCaffeinate
	startCaffeinate = func() (*exec.Cmd, error) {
		started++
		return &exec.Cmd{}, nil
	}
	stopCaffeinate = func(_ *exec.Cmd) error {
		stopped++
		return nil
	}
	t.Cleanup(func() {
		startCaffeinate = origStart
		stopCaffeinate = origStop
	})

	device := &mqtthass.Device{Name: "test-device"}

	worker, err := NewInhibitWorker(context.Background(), device)
	require.NoError(t, err)

	state, err := worker.stateCallback()
	require.NoError(t, err)
	assert.Equal(t, "OFF", string(state))

	worker.commandCallback(&paho.Publish{Payload: []byte("ON")})
	assert.Equal(t, 1, started)

	state, err = worker.stateCallback()
	require.NoError(t, err)
	assert.Equal(t, "ON", string(state))

	// Turning on again while already on should be a no-op.
	worker.commandCallback(&paho.Publish{Payload: []byte("ON")})
	assert.Equal(t, 1, started)

	worker.commandCallback(&paho.Publish{Payload: []byte("OFF")})
	assert.Equal(t, 1, stopped)

	state, err = worker.stateCallback()
	require.NoError(t, err)
	assert.Equal(t, "OFF", string(state))
}
