// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package dbusx

import (
	"log/slog"
)

const (
	LevelTrace = slog.Level(-8)
	LevelFatal = slog.Level(12)
)

func newLogger(bus string, parentLogger *slog.Logger) *slog.Logger {
	return parentLogger.With(slog.String("bus", bus))
}
