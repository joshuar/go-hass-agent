// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package device

import (
	"syscall"

	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/pkg/linux/whichdistro"
)

// getOSID will retrieve the distribution ID and version ID. These are
// suitable for usage as part of identifiers and variables. See also
// GetDistroDetails.
func getOSID() (id, versionid string) {
	var distroName, distroVersion string

	osReleaseInfo, err := whichdistro.GetOSRelease()
	if err != nil {
		log.Warn().Err(err).Msg("Could not read /etc/os-release. Contact your distro vendor to implement this file.")

		return unknownDistro, unknownDistroVersion
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

	return distroName, distroVersion
}

// GetOSDetails will retrieve the distribution name and version. The values
// are pretty-printed and may not be suitable for usage as identifiers and
// variables. See also GetDistroID.
func GetOSDetails() (name, version string) {
	var distroName, distroVersion string

	osReleaseInfo, err := whichdistro.GetOSRelease()
	if err != nil {
		log.Warn().Err(err).Msg("Could not read /etc/os-release. Contact your distro vendor to implement this file.")

		return unknownDistro, unknownDistroVersion
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

	return distroName, distroVersion
}

// GetKernelVersion will retrieve the kernel version.
//
//nolint:prealloc
func GetKernelVersion() string {
	var utsname syscall.Utsname

	var versionBytes []byte

	err := syscall.Uname(&utsname)
	if err != nil {
		log.Warn().Err(err).Msg("Could not retrieve kernel version.")

		return "Unknown"
	}

	for _, v := range utsname.Release {
		if v == 0 {
			continue
		}

		versionBytes = append(versionBytes, uint8(v))
	}

	return string(versionBytes)
}
