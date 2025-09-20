// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	UptimeFile = "/proc/uptime"
)

var (
	ErrDesktopPortalMissing = errors.New("no portal present")
	ErrUptimeInvalid        = errors.New("invalid uptime")
)

// findPortal is a helper function to work out which portal interface should be
// used for getting information on running apps.
func findPortal() (string, error) {
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

func getBootTime() (time.Time, error) {
	data, err := os.Open(UptimeFile)
	if err != nil {
		return time.Now(), fmt.Errorf("unable to read uptime: %w", err)
	}

	defer data.Close() //nolint:errcheck

	line := bufio.NewScanner(data)
	line.Split(bufio.ScanWords)

	if !line.Scan() {
		return time.Now(), ErrUptimeInvalid
	}

	uptimeValue, err := strconv.ParseFloat(line.Text(), 64)
	if err != nil {
		return time.Now(), ErrUptimeInvalid
	}

	uptime := time.Duration(uptimeValue * 1000000000) //nolint:mnd

	return time.Now().Add(-1 * uptime), nil
}
