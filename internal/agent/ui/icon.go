// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package ui

import _ "embed"

//go:embed assets/go-hass-agent.png
var hassIcon []byte

// TrayIcon satisfies the fyne.Resource interface to represent the tray icon.
type TrayIcon struct{}

func (i *TrayIcon) Name() string {
	return "TrayIcon"
}

func (i *TrayIcon) Content() []byte {
	return hassIcon
}
