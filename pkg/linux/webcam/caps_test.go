// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package webcam

import (
	"os"
	"testing"
)

func TestIsVideoCaptureDevice_liveCamera(t *testing.T) {
	if _, err := os.Stat("/dev/video0"); err != nil {
		t.Skip("no /dev/video0")
	}
	if _, err := os.Stat("/dev/video1"); err != nil {
		t.Skip("no /dev/video1")
	}

	InvalidateCaptureCache("")

	if !IsVideoCaptureDevice("/dev/video0") {
		t.Fatal("expected /dev/video0 to be a video capture device")
	}
	if IsVideoCaptureDevice("/dev/video1") {
		t.Fatal("expected /dev/video1 (UVC metadata node) to be excluded")
	}

	devs, err := GetVideoDevices()
	if err != nil {
		t.Fatalf("GetVideoDevices: %v", err)
	}
	for _, d := range devs {
		if d == "/dev/video1" {
			t.Fatalf("GetVideoDevices should not include metadata node, got %v", devs)
		}
	}
	if len(devs) == 0 {
		t.Fatal("expected at least /dev/video0")
	}
}
