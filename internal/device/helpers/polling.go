// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package helpers

import (
	"context"
	"sync"
	"time"

	"github.com/lthibault/jitterbug/v2"
)

// PollSensors is a helper function that will call the passed `updater()`
// function around each `interval` duration within the `stdev` duration window.
// Effectively, `updater()` will get called sometime near `interval`, but not
// exactly on it. This can help avoid a "thundering herd" problem of sensors all
// trying to update at the same time.
func PollSensors(ctx context.Context, updater func(), interval, stdev time.Duration) {
	updater()
	ticker := jitterbug.New(
		interval,
		&jitterbug.Norm{Stdev: stdev},
	)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				wg.Done()
				return
			case <-ticker.C:
				updater()
			}
		}
	}()
	wg.Wait()
}
