// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package linux

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"slices"
	"strconv"

	"kernel.org/pub/linux/libs/security/libcap/cap"
)

var ErrCapChecksFailed = errors.New("capability check failed")

// Checks contains system checks that are required to pass before a worker can start.
type Checks struct {
	// Groups is a list of group ids the user running Go Hass Agent needs to belong to.
	Groups []user.Group
	// Capabilities is the capabilities the Go Hass Agent binary needs.
	Capabilities []cap.Value
}

// Passed will perform all checks and return a boolean indicating whether they passed (true) or failed (false). On
// failure, on non-nil error will also be returned.
func (c *Checks) Passed() (bool, error) {
	groupsOK, err := c.hasGroups()
	if err != nil {
		return false, fmt.Errorf("%w: check for groups: %w", ErrCapChecksFailed, err)
	}
	if !groupsOK {
		return false, fmt.Errorf("%w: missing groups", ErrCapChecksFailed)
	}
	capsOK, err := c.hasCapabilities()
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrCapChecksFailed, err)
	}
	if !capsOK {
		return false, fmt.Errorf("%w: capabilities missing", ErrCapChecksFailed)
	}
	return true, nil
}

// hasGroups returns a boolean indicating whether Go Hass Agent is running with the required group permissions.
func (c *Checks) hasGroups() (bool, error) {
	gids, err := os.Getgroups()
	if err != nil {
		return false, fmt.Errorf("could not determine groups: %w", err)
	}
	for group := range slices.Values(c.Groups) {
		gid, err := strconv.Atoi(group.Gid)
		if err != nil {
			return false, fmt.Errorf("convert gid: %w", err)
		}
		if !slices.Contains(gids, gid) {
			return false, nil
		}
	}
	return true, nil
}

// hasCapabilities returns a boolean indicating whether Go Hass Agent has the required capabilties set.
func (c *Checks) hasCapabilities() (bool, error) {
	current := cap.GetProc()
	for required := range slices.Values(c.Capabilities) {
		found, err := current.GetFlag(cap.Permitted, required)
		if err != nil {
			return false, fmt.Errorf("could not parse required capability %s: %w", c.Capabilities, err)
		}
		if !found {
			return false, fmt.Errorf("missing %s", required.String())
		}
	}

	return true, nil
}
