// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"errors"
	"os"
	"strings"
)

var ErrDesktopPortalMissing = errors.New("no portal present")

// FindPortal is a helper function to work out which portal interface should be
// used for getting information on running apps.
func FindPortal() (string, error) {
	desktop := os.Getenv("XDG_CURRENT_DESKTOP")

	switch {
	case strings.Contains(desktop, "KDE"):
		return "org.freedesktop.impl.portal.desktop.kde", nil
	case strings.Contains(desktop, "GNOME"):
		return "org.freedesktop.impl.portal.desktop.gtk", nil
	default:
		return "", ErrDesktopPortalMissing
	}
}
