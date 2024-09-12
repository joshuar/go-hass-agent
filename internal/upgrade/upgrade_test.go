// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest
package upgrade

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

func TestRun(t *testing.T) {
	appID := "go-hass-agent-test"
	ctx := preferences.AppIDToContext(context.TODO(), appID)

	// Save the original value of paths for restoration after each test.
	oldPathOrig := oldPrefsPath
	oldRegistryPathOrig := oldRegistryPath
	type args struct {
		in0  context.Context
		path string
	}
	tests := []struct {
		args    args
		name    string
		wantErr bool
	}{
		{
			name: "with mqtt",
			args: args{in0: ctx, path: "testing/data/with-mqtt"},
		},
		{
			name: "without mqtt",
			args: args{in0: ctx, path: "testing/data/without-mqtt"},
		},
		{
			name:    "no previous preferences",
			args:    args{in0: ctx, path: "testing/data/nonexistent"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		oldPrefsPath = tt.args.path
		oldRegistryPath = filepath.Join(oldPrefsPath, "sensorRegistry")

		newPrefsPath = t.TempDir()
		xdg.ConfigHome = newPrefsPath
		newRegistryPath = filepath.Join(newPrefsPath, appID, "sensorRegistry")
		t.Run(tt.name, func(t *testing.T) {
			if err := Run(tt.args.in0); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
		if tt.wantErr {
			continue
		}

		var oldPrefs oldPreferences
		// Read the old preferences file.
		oldData, err := os.ReadFile(filepath.Join(oldPrefsPath, "preferences.toml"))
		require.NoError(t, err)
		err = toml.Unmarshal(oldData, &oldPrefs)
		require.NoError(t, err)
		// Read the new preferences file.
		newPrefs, err := preferences.Load(tt.args.in0)
		require.NoError(t, err)

		// Assert registration status have been preserved.
		assert.Equal(t, oldPrefs.Registered, newPrefs.Registered)
		// Assert the REST and Websocket APIs and the webhookid have been preserved.
		assert.Equal(t, oldPrefs.RestAPIURL, newPrefs.Hass.RestAPIURL)
		assert.Equal(t, oldPrefs.WebsocketURL, newPrefs.Hass.WebsocketURL)
		assert.Equal(t, oldPrefs.WebhookID, newPrefs.Hass.WebhookID)
		// Assert device ID and name has been preserved.
		assert.Equal(t, oldPrefs.DeviceID, newPrefs.Device.ID)
		assert.Equal(t, oldPrefs.DeviceName, newPrefs.Device.Name)
		// Assert MQTT status has been preserved.
		assert.Equal(t, oldPrefs.MQTTEnabled, newPrefs.MQTT.MQTTEnabled)
		// If MQTT enabled, assert the server has been preserved.
		if newPrefs.MQTT.MQTTEnabled {
			assert.Equal(t, oldPrefs.MQTTServer, newPrefs.MQTT.MQTTServer)
		}
		// Assert the registry file has been copied.
		assert.FileExists(t, filepath.Join(newRegistryPath, registryFile))

		oldPrefsPath = oldPathOrig
		oldRegistryPath = oldRegistryPathOrig
	}
}
