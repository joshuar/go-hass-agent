// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensors

import (
	"context"

	"fyne.io/fyne/v2"
)

type Registry interface {
	Open(ctx context.Context, registryPath fyne.URI) error
	Get(string) (*sensorMetadata, error)
	Set(string, *sensorMetadata) error
	Close() error
}
