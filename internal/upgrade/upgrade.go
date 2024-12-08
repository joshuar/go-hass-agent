// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package upgrade

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/pelletier/go-toml/v2"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

const (
	oldAppID     = "com.github.joshuar.go-hass-agent"
	registryFile = "sensor.reg"
)

var (
	oldPrefsPath    = filepath.Join(xdg.ConfigHome, oldAppID)
	oldRegistryPath = filepath.Join(xdg.ConfigHome, oldAppID, "sensorRegistry")

	ErrNoPrevConfig = errors.New("no directory from previous version found")
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

//nolint:cyclop
//revive:disable:function-length
func Run(ctx context.Context) error {
	newRegistryPath := filepath.Join(xdg.ConfigHome, preferences.AppIDFromContext(ctx), "sensorRegistry")

	// If there is no old preferences directory, exit.
	if _, err := os.Stat(oldPrefsPath); errors.Is(err, fs.ErrNotExist) {
		return errors.Join(ErrNoPrevConfig, err)
	}

	oldData, err := os.ReadFile(filepath.Join(oldPrefsPath, "preferences.toml"))
	if err != nil {
		return fmt.Errorf("cannot read old preferences: %w", err)
	}

	var oldPrefs oldPreferences

	if err = toml.Unmarshal(oldData, &oldPrefs); err != nil {
		return fmt.Errorf("cannot read old preferences: %w", err)
	}

	if err = preferences.Load(ctx); err != nil && !errors.Is(err, preferences.ErrLoadPreferences) {
		return fmt.Errorf("cannot initialize new preferences: %w", err)
	}

	// Registered status.
	if err = preferences.SetRegistered(oldPrefs.Registered); err != nil {
		return fmt.Errorf("cannot set new preferences: %w", err)
	}

	// MQTT preferences.
	if oldPrefs.MQTTEnabled {
		if err = preferences.SetMQTTPreferences(&preferences.MQTT{
			MQTTEnabled:     true,
			MQTTServer:      oldPrefs.MQTTServer,
			MQTTUser:        oldPrefs.MQTTUser,
			MQTTPassword:    oldPrefs.MQTTPassword,
			MQTTTopicPrefix: oldPrefs.MQTTTopicPrefix,
		}); err != nil {
			return fmt.Errorf("cannot set new preferences: %w", err)
		}
	}

	if err = preferences.SetHassPreferences(
		// Hass preferences.
		&preferences.Hass{
			Secret:       oldPrefs.Secret,
			WebhookID:    oldPrefs.WebhookID,
			CloudhookURL: oldPrefs.CloudhookURL,
			RemoteUIURL:  oldPrefs.RemoteUIURL,
			RestAPIURL:   oldPrefs.RestAPIURL,
			WebsocketURL: oldPrefs.WebsocketURL,
		},
		// Registration preferences.
		&preferences.Registration{
			Server: oldPrefs.Host,
			Token:  oldPrefs.Token,
		}); err != nil {
		return fmt.Errorf("cannot set new preferences: %w", err)
	}

	// Device preferences.
	if err = preferences.SetDevicePreferences(&preferences.Device{
		Name: oldPrefs.DeviceName,
		ID:   oldPrefs.DeviceID,
	}); err != nil {
		return fmt.Errorf("cannot set new preferences: %w", err)
	}

	err = preferences.Save(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", preferences.ErrSavePreferences, err)
	}

	// Create a directory for the registry.
	err = os.Mkdir(newRegistryPath, os.ModePerm)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("unable to create new registry directory: %w", err)
	}
	// If the directory already exists, we've been previously upgraded. Exit.
	if errors.Is(err, os.ErrExist) {
		return nil
	}

	// Attempt to copy over the old registry.
	oldReg, err := os.Open(filepath.Join(oldRegistryPath, registryFile))
	if err != nil {
		return fmt.Errorf("unable to open old registry: %w", err)
	}
	defer oldReg.Close()

	newReg, err := os.Create(filepath.Join(newRegistryPath, registryFile))
	if err != nil {
		return fmt.Errorf("unable to create new registry file: %w", err)
	}
	defer newReg.Close()

	_, err = io.Copy(newReg, oldReg)
	if err != nil {
		return fmt.Errorf("could not copy old registry contents: %w", err)
	}

	return nil
}
