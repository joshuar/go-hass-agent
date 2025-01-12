// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package device

import (
	"fmt"
	"syscall"

	"github.com/joshuar/go-hass-agent/pkg/linux/whichdistro"
)

// GetOSID will retrieve the distribution ID and version ID. These are
// suitable for usage as part of identifiers and variables. See also
// GetDistroDetails.
func GetOSID() (id, versionid string, err error) {
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
func GetOSDetails() (name, version string, err error) {
	var distroName, distroVersion string

	osReleaseInfo, err := whichdistro.GetOSRelease()
	if err != nil {
		return unknownDistro, unknownDistroVersion,
			fmt.Errorf("could not read /etc/os-release: %w", err)
	}

	if v, ok := osReleaseInfo.GetValue("NAME"); !ok {
		distroName = unknownDistro
	} else {
		distroName = v
	}

	if v, ok := osReleaseInfo.GetValue("VERSION"); !ok {
		distroVersion = unknownDistroVersion
	} else {
		distroVersion = v
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

		versionBytes = append(versionBytes, uint8(v))
	}

	return string(versionBytes), nil
}
