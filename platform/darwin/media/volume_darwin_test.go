package media

import (
	"context"
	"testing"

	"github.com/eclipse/paho.golang/paho"
	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVolumeWorkerStubbed(t *testing.T) {
	var ran []string

	origOutput, origRun := runOsascriptOutput, runOsascript
	runOsascriptOutput = func(_ context.Context, script string) (string, error) {
		ran = append(ran, script)
		switch script {
		case "output volume of (get volume settings)":
			return "42", nil
		case "output muted of (get volume settings)":
			return "true", nil
		}
		return "", nil
	}
	runOsascript = func(_ context.Context, script string) error {
		ran = append(ran, script)
		return nil
	}
	t.Cleanup(func() {
		runOsascriptOutput = origOutput
		runOsascript = origRun
	})

	device := &mqtthass.Device{Name: "test-device"}
	ctx := context.Background()

	worker, err := NewVolumeWorker(ctx, device)
	require.NoError(t, err)
	require.NotNil(t, worker.VolumeControl)
	require.NotNil(t, worker.MuteControl)

	state, err := worker.volumeStateCallback(ctx)
	require.NoError(t, err)
	assert.Equal(t, `{"value": 42}`, string(state))

	state, err = worker.muteStateCallback(ctx)
	require.NoError(t, err)
	assert.Equal(t, "ON", string(state))

	worker.volumeCommandCallback(ctx, &paho.Publish{Payload: []byte("80")})
	assert.Contains(t, ran, "set volume output volume 80")

	worker.muteCommandCallback(ctx, &paho.Publish{Payload: []byte("OFF")})
	assert.Contains(t, ran, "set volume without output muted")
}

// TestVolumeReal exercises the real osascript read path against this
// machine's actual volume/mute state. It only reads, plus one write-back of
// the volume to its own current value (verified manually to be a no-op),
// never an actual mute toggle.
func TestVolumeReal(t *testing.T) {
	ctx := context.Background()

	vol, err := runOsascriptOutput(ctx, "output volume of (get volume settings)")
	require.NoError(t, err)

	require.NoError(t, runOsascript(ctx, "set volume output volume "+vol))

	muted, err := runOsascriptOutput(ctx, "output muted of (get volume settings)")
	require.NoError(t, err)
	assert.Contains(t, []string{"true", "false"}, muted)
}
