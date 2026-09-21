package device

import (
	"fmt"
	"strings"

	"golang.org/x/sys/unix"
)

// GetOSID will retrieve the distribution ID and version ID. These are
// suitable for usage as part of identifiers and variables. See also
// GetDistroDetails.
func GetOSID() (string, string, error) {
	var uts unix.Utsname

	if err := unix.Uname(&uts); err != nil {
		return unknownDistro, unknownDistroVersion,
			fmt.Errorf("could not retrieve kernel version: %w", err)
	}

	// Release is a fixed-size buffer padded with NUL bytes, cut it at the
	// first one.
	distroVersion, _, _ := strings.Cut(string(uts.Release[:]), "\x00")

	return "macos", distroVersion, nil
}
