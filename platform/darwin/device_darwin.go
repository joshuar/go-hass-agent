package darwin

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

func getBootTime() (time.Time, error) {
	out, err := unix.SysctlRaw("kern.boottime")
	if err != nil {
		return time.Now(), err
	}
	var timeval syscall.Timeval
	if len(out) != int(unsafe.Sizeof(timeval)) {
		return time.Now(), fmt.Errorf("unexpected output of sysctl kern.boottime: %v (len: %d)", out, len(out))
	}

	timeval = *(*syscall.Timeval)(unsafe.Pointer(&out[0]))
	sec, nsec := timeval.Unix()

	return time.Unix(sec, nsec), nil
}
