// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"time"

	"github.com/lthibault/jitterbug/v2"
)

func pollSensors(ctx context.Context, updater func(), interval time.Duration, stdev time.Duration) {
	updater()
	ticker := jitterbug.New(
		interval,
		&jitterbug.Norm{Stdev: stdev},
	)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				updater()
			}
		}
	}()
}
