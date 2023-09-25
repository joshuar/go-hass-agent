// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package device

import (
	"context"
	"sync"
)

// SensorTracker is the interface through which a device can update its sensors.
type SensorTracker interface {
	// UpdateSensors will take any number of sensor updates and pass them on to
	// the sensor tracker, which will handle updating its internal database and
	// Home Assistant.
	UpdateSensors(context.Context, ...interface{}) error
}

// StartWorkers will call all the sensor worker functions that have been defined
// for this device.
func StartWorkers(ctx context.Context, workers []func(context.Context, SensorTracker), tracker SensorTracker) {
	var wg sync.WaitGroup
	for _, worker := range workers {
		wg.Add(1)
		go func(worker func(context.Context, SensorTracker)) {
			defer wg.Done()
			worker(ctx, tracker)
		}(worker)
	}
	wg.Wait()
}
