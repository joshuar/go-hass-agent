// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package info

import (
	"fmt"
	"syscall"

	"github.com/joshuar/go-hass-agent/pkg/linux/whichdistro"
)

// GetOSID will retrieve the distribution ID and version ID. These are
// suitable for usage as part of identifiers and variables. See also
// GetDistroDetails.
func GetOSID() (string, string, error) {
	var distroName, distroVersion string

	osReleaseInfo, err := whichdistro.GetOSRelease()
	if err != nil {
		return unknownDistro, unknownDistroVersion,
			fmt.Errorf("could not read /etc/os-release: %w", err)
	}

	if v, ok := osReleaseInfo.GetValue("ID"); !ok {
		distroName = unknownDistro
	} else {
		distroName = v
	}

	if v, ok := osReleaseInfo.GetValue("VERSION_ID"); !ok {
		distroVersion = unknownDistroVersion
	} else {
		distroVersion = v
	}

	return distroName, distroVersion, nil
}

// GetOSDetails will retrieve the distribution name and version. The values
// are pretty-printed and may not be suitable for usage as identifiers and
// variables. See also GetDistroID.
func GetOSDetails() (string, string, error) {
	var (
		distroName, distroVersion string
		value                     string
		found                     bool
	)

	osReleaseInfo, err := whichdistro.GetOSRelease()
	if err != nil {
		return unknownDistro, unknownDistroVersion,
			fmt.Errorf("could not read /etc/os-release: %w", err)
	}

	if value, found = osReleaseInfo.GetValue("NAME"); found {
		distroName = value
	} else if value, found = osReleaseInfo.GetValue("ID"); found {
		distroName = value
	} else {
		distroName = unknownDistro
	}

	if value, found = osReleaseInfo.GetValue("VERSION"); found {
		distroVersion = value
	} else if value, found = osReleaseInfo.GetValue("VERSION_ID"); found {
		distroVersion = value
	} else {
		distroVersion = unknownDistroVersion
	}

	return distroName, distroVersion, nil
}

// GetKernelVersion will retrieve the kernel version.
//
//nolint:prealloc
func GetKernelVersion() (string, error) {
	var utsname syscall.Utsname

	var versionBytes []byte

	err := syscall.Uname(&utsname)
	if err != nil {
		return unknownValue, fmt.Errorf("could not retrieve kernel version: %w", err)
	}

	for _, v := range utsname.Release {
		if v == 0 {
			continue
		}

		versionBytes = append(versionBytes, uint8(v)) // #nosec: G115
	}

	return string(versionBytes), nil
}
