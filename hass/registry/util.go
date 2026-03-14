// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package registry

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

func checkPath(path string) error {
	if _, err := os.Stat(path); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("stat directory: %w", err)
		}
		if err := os.MkdirAll(path, 0700); err != nil {
			return fmt.Errorf("create directory %s: %w", path, err)
		}
	}
	return nil
}
