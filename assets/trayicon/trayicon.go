// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package trayicon

type TrayIcon struct{}

func (icon *TrayIcon) Name() string {
	return "TrayIcon"
}

func (icon *TrayIcon) Content() []byte {
	return home_assistant_icon
}
