// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
)

//go:generate moq -out mockDevice.go . Device
type Device interface {
	DeviceName() string
	DeviceID() string
	Setup(context.Context) context.Context
}
