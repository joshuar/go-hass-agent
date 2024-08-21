// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest
package agent

import (
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

func TestAgent_newMQTTDevice(t *testing.T) {
	hostname, err := device.GetHostname(true)
	require.NoError(t, err)

	prefs := preferences.DefaultPreferences(filepath.Join(t.TempDir(), "test.toml"))

	identifiers := []string{"go-hass-agent-test", prefs.Device.Name, prefs.Device.ID}

	type fields struct {
		ui            UI
		hass          HassClient
		done          chan struct{}
		prefs         *preferences.Preferences
		logger        *slog.Logger
		id            string
		headless      bool
		forceRegister bool
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "valid device",
			fields: fields{
				prefs:  prefs,
				id:     "go-hass-agent-test",
				logger: slog.Default(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				ui:            tt.fields.ui,
				hass:          tt.fields.hass,
				done:          tt.fields.done,
				prefs:         tt.fields.prefs,
				logger:        tt.fields.logger,
				id:            tt.fields.id,
				headless:      tt.fields.headless,
				forceRegister: tt.fields.forceRegister,
			}
			got := agent.newMQTTDevice()
			// Assert the MQTT device name is the device hostname.
			assert.Equal(t, hostname, got.Name)
			// Assert the MQTT device identifiers are the expected values.
			assert.Equal(t, identifiers, got.Identifiers)
		})
	}
}
