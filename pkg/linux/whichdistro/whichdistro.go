// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Package whichdistro provides methods to determine the Linux distribution from the os-release file.
package whichdistro

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	// UnknownValue is used when the value could not otherwise be determined.
	UnknownValue = "Unknown"
)

var (
	// OSReleaseFile is the canonical location of the os-release file.
	OSReleaseFile = "/etc/os-release"
	// OSReleaseAltFile is a canonical alternative location of the os-release file.
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
		if len(fields) == 2 {
			info[string(fields[0])] = string(fields[1])
		}
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

	return nil, fmt.Errorf("unable to read OSRelease file: %w", err)
}

// GetValue will retrieve the value of the given key from an OSRelease map. It
// will perform some cleanup on the raw value to make it easier to use.
func (r OSRelease) GetValue(key string) (string, bool) {
	value, ok := r[key]
	if !ok {
		return UnknownValue, false
	}

	if strings.ContainsAny(value, `"`) {
		unquoted, err := strconv.Unquote(value)
		if err != nil {
			return UnknownValue, false
		}

		value = unquoted
	}

	return value, true
}
