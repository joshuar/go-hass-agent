// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package linux

import (
	"time"
)

type RateValue[T ~uint64] struct {
	prevValue T
}

func (r *RateValue[T]) Calculate(currValue T, delta time.Duration) T {
	var rate T

	if T(delta.Seconds()) > 0 {
		rate = ((currValue - r.prevValue) / T(delta.Seconds()))
	}

	r.prevValue = currValue

	return rate
}
