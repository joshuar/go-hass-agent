// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package whichdistro

import (
	"bytes"
	"os"
)

var (
	OSReleaseFile    = "/etc/os-release"
	OSReleaseAltFile = "/usr/lib/os-release"
)

// GetOSRelease will fetch the OS Release info from the canonical file
// locations. The data will be formatted as a map[string]string. If the OS
// Release info cannot be read, an error will be returned containing details of
// why.
func GetOSRelease() (map[string]string, error) {
	info := make(map[string]string)
	file, err := readOSRelease()
	if err != nil {
		return nil, err
	}
	lines := bytes.Split(file, []byte("\n"))
	for _, line := range lines {
		if bytes.Equal(line, []byte("")) {
			continue
		}
		fields := bytes.FieldsFunc(line, func(r rune) bool {
			return r == '='
		})
		info[string(fields[0])] = string(fields[1])
	}
	return info, nil
}

func readOSRelease() ([]byte, error) {
	var contents []byte
	var err error
	contents, err = os.ReadFile(OSReleaseFile)
	if err == nil {
		return contents, nil
	}
	contents, err = os.ReadFile(OSReleaseAltFile)
	if err == nil {
		return contents, nil
	}
	return nil, err
}
