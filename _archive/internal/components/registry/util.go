// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package registry

import (
	"fmt"
	"os"
)

func checkPath(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err := os.MkdirAll(path, 0o750)
		if err != nil {
			return fmt.Errorf("unable to create new directory: %w", err)
		}
	}

	return nil
}
