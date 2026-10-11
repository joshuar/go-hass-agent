package power

import (
	"context"
	"log/slog"
	"sync"
	"testing"

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPowerControl(t *testing.T) {
	var (
		mu      sync.Mutex
		scripts []string
	)

	orig := runOsascript
	runOsascript = func(_ context.Context, script string) error {
		mu.Lock()
		defer mu.Unlock()
		scripts = append(scripts, script)
		return nil
	}
	t.Cleanup(func() { runOsascript = orig })

	device := &mqtthass.Device{Name: "test-device"}

	buttons, err := NewPowerControl(context.Background(), device)
	require.NoError(t, err)
	require.Len(t, buttons, len(powerCommands))

	ids := make(map[string]bool)
	for _, b := range buttons {
		ids[b.UniqueID] = true
	}
	for _, c := range powerCommands {
		assert.True(t, ids["test-device_"+c.idSfx], "expected a button for %s", c.name)
	}

	// Press each command and confirm it runs its own script -- never
	// actually sleeping/restarting/shutting down this test machine, since
	// runOsascript is stubbed above.
	logger := slog.Default()
	for _, c := range powerCommands {
		pressPowerCommand(context.Background(), logger, c)
		assert.Contains(t, scripts, c.script)
	}
}
