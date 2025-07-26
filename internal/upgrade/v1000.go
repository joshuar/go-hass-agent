// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package upgrade

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
)

const (
	oldAppID     = "com.github.joshuar.go-hass-agent"
	registryFile = "sensor.reg"
)

type oldPreferences struct {
	WebhookID       string `toml:"hass.webhookid"`
	RemoteUIURL     string `toml:"hass.remoteuiurl,omitempty"`
	DeviceName      string `toml:"device.name"`
	MQTTTopicPrefix string `toml:"mqtt.topic_prefix,omitempty"`
	Host            string `toml:"registration.host"`
	RestAPIURL      string `toml:"hass.apiurl,omitempty"`
	Token           string `toml:"registration.token"`
	CloudhookURL    string `toml:"hass.cloudhookurl,omitempty"`
	MQTTPassword    string `toml:"mqtt.password,omitempty"`
	Version         string `toml:"agent.version"`
	DeviceID        string `toml:"device.id"`
	WebsocketURL    string `toml:"hass.websocketurl"`
	MQTTServer      string `toml:"mqtt.server,omitempty"`
	MQTTUser        string `toml:"mqtt.user,omitempty"`
	Secret          string `toml:"hass.secret,omitempty"`
	MQTTEnabled     bool   `toml:"mqtt.enabled"`
	Registered      bool   `toml:"hass.registered"`
}

//nolint:cyclop,funlen
func v1000(ctx context.Context) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("cannot read user config dir: %w", err)
	}
	oldPrefsPath := filepath.Join(configDir, oldAppID)
	oldRegistryPath := filepath.Join(configDir, oldAppID, "sensorRegistry")

	newRegistryPath := filepath.Join(preferences.PathFromCtx(ctx), "sensorRegistry")

	// If there is no old preferences directory, exit.
	if _, err := os.Stat(oldPrefsPath); errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	oldData, err := os.ReadFile(filepath.Join(oldPrefsPath, "preferences.toml")) // #nosec: G304
	if err != nil {
		return fmt.Errorf("cannot read old preferences: %w", err)
	}

	var oldPrefs oldPreferences

	if err = toml.Unmarshal(oldData, &oldPrefs); err != nil {
		return fmt.Errorf("cannot read old preferences: %w", err)
	}

	if err = preferences.Init(ctx); err != nil && !errors.Is(err, preferences.ErrLoadPreferences) {
		return fmt.Errorf("cannot initialize new preferences: %w", err)
	}

	var preferencesToSet []preferences.SetPreference

	// Set MQTT preferences if MQTT is enabled.
	if oldPrefs.MQTTEnabled {
		preferencesToSet = append(preferencesToSet,
			preferences.SetMQTTEnabled(true),
			preferences.SetMQTTServer(oldPrefs.MQTTServer),
			preferences.SetMQTTTopicPrefix(oldPrefs.MQTTTopicPrefix),
			preferences.SetMQTTUser(oldPrefs.MQTTUser),
			preferences.SetMQTTPassword(oldPrefs.MQTTPassword),
		)
	}
	// Set all other required preferences.
	preferencesToSet = append(preferencesToSet,
		// Hass preferences.
		preferences.SetHassSecret(oldPrefs.Secret),
		preferences.SetRestAPIURL(oldPrefs.RestAPIURL),
		preferences.SetWebsocketURL(oldPrefs.WebsocketURL),
		preferences.SetWebhookID(oldPrefs.WebhookID),
		// Device preferences.
		preferences.SetDeviceID(oldPrefs.DeviceID),
		preferences.SetDeviceName(oldPrefs.DeviceName),
		preferences.SetRegistered(true),
	)

	if err = preferences.Set(preferencesToSet...); err != nil {
		return fmt.Errorf("%w: %w", preferences.ErrSavePreferences, err)
	}

	// Create a directory for the registry.
	err = os.Mkdir(newRegistryPath, 0o750)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("unable to create new registry directory: %w", err)
	}
	// If the directory already exists, we've been previously upgraded. Exit.
	if errors.Is(err, os.ErrExist) {
		return nil
	}

	// Attempt to copy over the old registry.
	oldReg, err := os.Open(filepath.Join(oldRegistryPath, registryFile)) // #nosec: G304
	if err != nil {
		return fmt.Errorf("unable to open old registry: %w", err)
	}
	defer oldReg.Close() //nolint:errcheck

	newReg, err := os.Create(filepath.Join(newRegistryPath, registryFile)) // #nosec: G304
	if err != nil {
		return fmt.Errorf("unable to create new registry file: %w", err)
	}
	defer newReg.Close() //nolint:errcheck

	_, err = io.Copy(newReg, oldReg)
	if err != nil {
		return fmt.Errorf("could not copy old registry contents: %w", err)
	}

	return nil
}
