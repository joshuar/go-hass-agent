// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package whichdistro

import (
	"bytes"
	"os"
	"strconv"
)

const (
	UnknownValue = "Unknown"
)

var (
	OSReleaseFile    = "/etc/os-release"
	OSReleaseAltFile = "/usr/lib/os-release"
)

// OSRelease is a map of the OS Release file keys and values. See the
// os-release(5) manpage for information on what keys and their values might be
// available.
type OSRelease map[string]string

// GetOSRelease will fetch the OS Release info from the canonical file
// locations. If the OS Release info cannot be read, an error will be returned
// containing details of why.
func GetOSRelease() (OSRelease, error) {
	info := make(OSRelease)
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

// GetValue will retrieve the value of the given key from an OSRelease map. It
// will perform some cleanup on the raw value to make it easier to use.
func (r OSRelease) GetValue(key string) (value string, ok bool) {
	if v, ok := r[key]; !ok {
		return UnknownValue, false
	} else {
		value, err := strconv.Unquote(v)
		if err != nil {
			return UnknownValue, false
		}
		return value, true
	}
}
