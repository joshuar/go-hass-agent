// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import "github.com/joshuar/go-hass-agent/pkg/linux/hwmon"

func main() {
	for _, s := range hwmon.GetAllSensors() {
		println(s.String())
	}
}
