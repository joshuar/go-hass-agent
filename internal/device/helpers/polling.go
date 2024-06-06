// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package helpers

import (
	"context"
	"time"

	"github.com/lthibault/jitterbug/v2"
)

// PollSensors is a helper function that will call the passed `updater()`
// function around each `interval` duration within the `stdev` duration window.
// Effectively, `updater()` will get called sometime near `interval`, but not
// exactly on it. This can help avoid a "thundering herd" problem of sensors all
// trying to update at the same time.
func PollSensors(ctx context.Context, updater func(time.Duration), interval, stdev time.Duration) {
	lastTick := time.Now()
	updater(time.Since(lastTick))
	ticker := jitterbug.New(
		interval,
		&jitterbug.Norm{Stdev: stdev},
	)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case t := <-ticker.C:
			updater(time.Since(lastTick))
			lastTick = t
		}
	}
}
