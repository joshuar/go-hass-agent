// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package device

import "context"

// SensorTracker is the interface through which a device can update its sensors.
type SensorTracker interface {
	// UpdateSensors will take any number of sensor updates and pass them on to
	// the sensor tracker, which will handle updating its internal database and
	// Home Assistant.
	UpdateSensors(context.Context, ...interface{}) error
}
