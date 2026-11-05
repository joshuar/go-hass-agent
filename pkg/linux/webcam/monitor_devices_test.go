// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package webcam

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetVideoDevicesGlobIsVideoOnly(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	for _, name := range []string{"video0", "video1", "null", "urandom", "tty", "media0"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o600); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
	}

	// Mirrors the Glob used by GetVideoDevices so non-video /dev nodes are excluded.
	matches, err := filepath.Glob(filepath.Join(dir, "video*"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("expected only video0/video1, got %v", matches)
	}
}
