// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package ui

import _ "embed"

//go:embed assets/go-hass-agent.png
var hassIcon []byte

// trayIcon satisfies the fyne.Resource interface to represent the tray icon.
type trayIcon struct{}

func (i *trayIcon) Name() string {
	return "TrayIcon"
}

func (i *trayIcon) Content() []byte {
	return hassIcon
}
