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
func StartWorkers(ctx context.Context, tracker SensorTracker, workers ...func(context.Context, SensorTracker)) {
	workerCh := make(chan func(context.Context, SensorTracker), len(workers))

	for i := 0; i < len(workerCh); i++ {
		workerCh <- workers[i]
	}

	var wg sync.WaitGroup
	for _, workerFunc := range workers {
		wg.Add(1)
		go func(workerFunc func(context.Context, SensorTracker)) {
			defer wg.Done()
			workerFunc(ctx, tracker)
		}(workerFunc)
	}
	
	close(workerCh)
	wg.Wait()
}
