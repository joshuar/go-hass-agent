// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package webcam

import (
	"fmt"
	"os"
	"sync"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	v4l2CapVideoCapture = 0x00000001
	v4l2CapDeviceCaps   = 0x80000000
)

// v4l2Capability mirrors linux/videodev2.h struct v4l2_capability.
type v4l2Capability struct {
	Driver       [16]byte
	Card         [32]byte
	BusInfo      [32]byte
	Version      uint32
	Capabilities uint32
	DeviceCaps   uint32
	Reserved     [3]uint32
}

// VIDIOC_QUERYCAP is _IOR('V', 0, struct v4l2_capability).
var vidIOCQueryCap = ioctlIOR('V', 0, unsafe.Sizeof(v4l2Capability{}))

var (
	captureCacheMu sync.Mutex
	captureCache   = map[string]bool{}
)

func ioctlIOR(t byte, nr byte, size uintptr) uintptr {
	const (
		iocNRBits   = 8
		iocTypeBits = 8
		iocSizeBits = 14
		iocDirBits  = 2

		iocNRShift   = 0
		iocTypeShift = iocNRShift + iocNRBits
		iocSizeShift = iocTypeShift + iocTypeBits
		iocDirShift  = iocSizeShift + iocSizeBits

		iocRead = 2
	)
	return uintptr(iocRead)<<iocDirShift |
		size<<iocSizeShift |
		uintptr(t)<<iocTypeShift |
		uintptr(nr)<<iocNRShift
}

// IsVideoCaptureDevice reports whether path is a V4L2 node that can capture
// video. Metadata-only UVC nodes (common as /dev/video1) return false.
// Results are cached so QUERYCAP opens do not keep re-probing the device.
func IsVideoCaptureDevice(path string) bool {
	captureCacheMu.Lock()
	if cached, ok := captureCache[path]; ok {
		captureCacheMu.Unlock()
		return cached
	}
	captureCacheMu.Unlock()

	ok, err := queryVideoCapture(path)
	if err != nil {
		// If QUERYCAP fails, keep the device so real cameras are not dropped
		// on transient errors; do not cache failures.
		return true
	}

	captureCacheMu.Lock()
	captureCache[path] = ok
	captureCacheMu.Unlock()
	return ok
}

// InvalidateCaptureCache clears cached capability results (e.g. on hotplug).
func InvalidateCaptureCache(path string) {
	captureCacheMu.Lock()
	defer captureCacheMu.Unlock()
	if path == "" {
		captureCache = map[string]bool{}
		return
	}
	delete(captureCache, path)
}

func queryVideoCapture(path string) (bool, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return false, fmt.Errorf("open device: %w", err)
	}
	defer f.Close()

	var caps v4l2Capability
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		f.Fd(),
		vidIOCQueryCap,
		uintptr(unsafe.Pointer(&caps)),
	)
	if errno != 0 {
		return false, errno
	}

	deviceCaps := caps.Capabilities
	if caps.Capabilities&v4l2CapDeviceCaps != 0 {
		deviceCaps = caps.DeviceCaps
	}
	return deviceCaps&v4l2CapVideoCapture != 0, nil
}
