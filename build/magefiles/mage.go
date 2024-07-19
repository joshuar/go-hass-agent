// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"log/slog"
	"os"
)

func init() {
	// set global logger with custom options
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
}
