// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

// revive:disable:unused-receiver

package device

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/iancoleman/strcase"
	"github.com/jaypipes/ghw"
	mqtthass "github.com/joshuar/go-hass-anything/v9/pkg/hass"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

const (
	unknownVendor        = "Unknown Vendor"
	unknownModel         = "Unknown Model"
	unknownDistro        = "Unknown Distro"
	unknownDistroVersion = "Unknown Version"
	UnknownValue         = "unknown"
)

var ErrUnsupportedHardware = errors.New("unsupported hardware")

// New creates a new hass.DeviceInfo based on the device running this agent.
// Note that the device is not idempotent, each call to this function will have
// at least a different DeviceID in addition to any other non-static variables
// such as the hostname.
//
//nolint:exhaustruct
func New(ctx context.Context, name, version string) *hass.DeviceInfo {
	hostname, err := getHostname(true)
	if err != nil {
		logging.FromContext(ctx).Warn("Problem occurred.", "error", err.Error())
	}

	deviceID, err := getDeviceID()
	if err != nil {
		logging.FromContext(ctx).Warn("Problem occurred.", "error", err.Error())
	}

	dev := &hass.DeviceInfo{
		AppName:            name,
		AppVersion:         version,
		AppID:              strcase.ToSnake(name),
		DeviceID:           deviceID,
		DeviceName:         hostname,
		SupportsEncryption: false,
		AppData: hass.AppData{
			PushWebsocket: true,
		},
	}

	dev.OsName, dev.OsVersion, err = getOSID()
	if err != nil {
		logging.FromContext(ctx).Warn("Problem occurred.", "error", err.Error())
	}

	dev.Model, dev.Manufacturer, err = getHWProductInfo()
	if err != nil {
		logging.FromContext(ctx).Warn("Problem occurred.", "error", err.Error())
	}

	return dev
}

// MQTTDeviceInfo returns an mqtthas.Device with the required info for
// representing the device running the agent in MQTT.
//
//nolint:exhaustruct
func MQTTDeviceInfo(ctx context.Context) *mqtthass.Device {
	prefs, err := preferences.ContextGetPrefs(ctx)
	if err != nil {
		logging.FromContext(ctx).Warn("Could not retrieve preferences.", "error", err.Error())
	}

	hostname, err := getHostname(true)
	if err != nil {
		logging.FromContext(ctx).Warn("Problem occurred.", "error", err.Error())
	}

	_, version, err := getOSID()
	if err != nil {
		logging.FromContext(ctx).Warn("Problem occurred.", "error", err.Error())
	}

	model, manufacturer, err := getHWProductInfo()
	if err != nil {
		logging.FromContext(ctx).Warn("Problem occurred.", "error", err.Error())
	}

	return &mqtthass.Device{
		Name:         hostname,
		URL:          preferences.AppURL,
		SWVersion:    version,
		Manufacturer: manufacturer,
		Model:        model,
		Identifiers:  []string{prefs.DeviceID},
	}
}

// getDeviceID create a new device ID. It will be a randomly generated UUIDv4.
func getDeviceID() (string, error) {
	deviceID, err := uuid.NewV4()
	if err != nil {
		return UnknownValue, fmt.Errorf("could not retrieve a machine ID: %w", err)
	}

	return deviceID.String(), nil
}

// getHostname retrieves the hostname of the device running the agent, or
// localhost if that doesn't work.
//
//revive:disable:flag-parameter
func getHostname(short bool) (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "localhost", fmt.Errorf("could not retrieve hostname: %w", err)
	}

	if short {
		shortHostname, _, _ := strings.Cut(hostname, ".")

		return shortHostname, nil
	}

	return hostname, nil
}

// getHWProductInfo retrieves the model and vendor of the machine. If these
// cannot be retrieved or cannot be found, they will be set to default unknown
// strings.
func getHWProductInfo() (model, vendor string, err error) {
	product, err := ghw.Product(ghw.WithDisableWarnings())
	if err != nil {
		return unknownModel, unknownVendor, fmt.Errorf("could not retrieve hardware information: %w", err)
	}

	return product.Name, product.Vendor, nil
}

// Chassis will return the chassis type of the machine, such as "desktop" or
// "laptop". If this cannot be retrieved, it will return "unknown".
func Chassis() (string, error) {
	chassisInfo, err := ghw.Chassis(ghw.WithDisableWarnings())
	if err != nil {
		return UnknownValue, fmt.Errorf("could not determine chassis type: %w", err)
	}

	return chassisInfo.Type, nil
}
