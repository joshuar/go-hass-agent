// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

// revive:disable:unused-receiver
package device

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/jaypipes/ghw"
)

const (
	unknownVendor        = "Unknown Vendor"
	unknownModel         = "Unknown Model"
	unknownDistro        = "Unknown Distro"
	unknownDistroVersion = "Unknown Version"
	unknownValue         = "unknown"
	defaultHostname      = "localhost"
)

var ErrUnsupportedHardware = errors.New("unsupported hardware")

// Chassis will return the chassis type of the machine, such as "desktop" or
// "laptop". If this cannot be retrieved, it will return "unknown".
func Chassis() (string, error) {
	chassisInfo, err := ghw.Chassis(ghw.WithDisableWarnings())
	if err != nil || chassisInfo == nil {
		return unknownValue, fmt.Errorf("could not determine chassis type: %w", err)
	}

	return chassisInfo.TypeDescription, nil
}

// GetHostname retrieves the hostname of the device running the agent, or
// localhost if that doesn't work.
func GetHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return defaultHostname, fmt.Errorf("could not retrieve hostname: %w", err)
	}

	shortHostname, _, _ := strings.Cut(hostname, ".")

	return shortHostname, nil
}

// GetHWProductInfo retrieves the model and vendor of the machine. If these
// cannot be retrieved or cannot be found, they will be set to default unknown
// strings.
func GetHWProductInfo() (model, vendor string, err error) {
	product, err := ghw.Product(ghw.WithDisableWarnings())
	if err != nil || product == nil {
		return unknownModel, unknownVendor, fmt.Errorf("could not retrieve hardware information: %w", err)
	}

	return product.Name, product.Vendor, nil
}

// NewDeviceID create a new device ID. It will be a randomly generated UUIDv4.
func NewDeviceID() (string, error) {
	deviceID, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("could not retrieve a machine ID: %w", err)
	}

	return deviceID.String(), nil
}
