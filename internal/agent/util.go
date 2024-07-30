// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/joshuar/go-hass-agent/internal/scripts"
)

// mergeCh merges a list of channels into a single channel. Essentially, fan-in for
// channels.
func mergeCh[T any](ctx context.Context, inCh ...<-chan T) chan T {
	var wg sync.WaitGroup

	outCh := make(chan T)

	// Start an output goroutine for each input channel in sensorCh.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(ch <-chan T) { //nolint:varnamelen
		defer wg.Done()

		if ch == nil {
			return
		}

		for n := range ch {
			select {
			case outCh <- n:
			case <-ctx.Done():
				return
			}
		}
	}

	wg.Add(len(inCh))

	for _, c := range inCh {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(outCh)
	}()

	return outCh
}

// FindScripts locates scripts and returns a slice of scripts that the agent can
// run.
func findScripts(path string) ([]Script, error) {
	var sensorScripts []Script

	var errs error

	files, err := filepath.Glob(path + "/*")
	if err != nil {
		return nil, fmt.Errorf("could not search for scripts: %w", err)
	}

	for _, scriptFile := range files {
		if isExecutable(scriptFile) {
			script, err := scripts.NewScript(scriptFile)
			if err != nil {
				errs = errors.Join(errs, err)

				continue
			}

			sensorScripts = append(sensorScripts, script)
		}
	}

	return sensorScripts, nil
}

func isExecutable(filename string) bool {
	fi, err := os.Stat(filename)
	if err != nil {
		return false
	}

	return fi.Mode().Perm()&0o111 != 0
}
