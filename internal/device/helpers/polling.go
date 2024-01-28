// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
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
func PollSensors(ctx context.Context, updater func(time.Duration), interval, stdev time.Duration) {
	var wg sync.WaitGroup
	lastTick := time.Now()
	wg.Add(1)
	go func() {
		defer wg.Done()
		updater(time.Since(lastTick))
	}()
	wg.Wait()
	ticker := jitterbug.New(
		interval,
		&jitterbug.Norm{Stdev: stdev},
	)
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			wg.Add(1)
			go func() {
				defer wg.Done()
				updater(time.Since(lastTick))
			}()
			wg.Wait()
			lastTick = t
		}
	}
}
