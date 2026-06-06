// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package webcam

import (
	"context"
	"encoding/binary"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/platform/linux"
)

const (
	inotifyAccess     = 0x00000001 // IN_ACCESS
	inotifyModify     = 0x00000002 // IN_MODIFY
	inotifyCloseWrite = 0x00000008 // IN_CLOSE_WRITE
	inotifyCloseNoWrt = 0x00000010 // IN_CLOSE_NOWRITE  (read-only close)
	inotifyOpen       = 0x00000020 // IN_OPEN
	inotifyCreate     = 0x00000100 // IN_CREATE
	inotifyDelete     = 0x00000200 // IN_DELETE
	inotifyMovedTo    = 0x00000080 // IN_MOVED_TO
	inotifyMovedFrom  = 0x00000040 // IN_MOVED_FROM
	inotifyDeleteSelf = 0x00000400 // IN_DELETE_SELF
	inotifyMoveself   = 0x00000800 // IN_MOVE_SELF
	inotifyOnlydir    = 0x01000000 // IN_ONLYDIR
	inotifyDontFollow = 0x02000000 // IN_DONT_FOLLOW
	inotifyExclUnlink = 0x04000000 // IN_EXCL_UNLINK
	inotifyMaskAdd    = 0x20000000 // IN_MASK_ADD
	inotifyIsdir      = 0x40000000 // IN_ISDIR
	inotifyOneshot    = 0x80000000 // IN_ONESHOT

	// Combined mask we actually watch for.
	watchMask = inotifyOpen | inotifyCloseWrite | inotifyCloseNoWrt | inotifyCreate | inotifyDelete
)

const (
	// StatusActive indicates a webcam is active.
	StatusActive Status = "ACTIVE"
	// StatusIdle indicates a webcam is idle (i.e., off/inactive).
	StatusIdle Status = "IDLE"
)

// Status indicates the status of a webcam.
type Status string

// Event is an event triggered from a webcam.
type Event struct {
	Timestamp time.Time
	Device    string
	Status    Status
	Message   string
}

// Monitor tracks the state of video device usage.
type Monitor struct {
	mu      sync.Mutex
	active  map[string]int // device path → open-count estimate
	eventCh chan Event
}

var (
	// DefaultReconcilerInterval is the interval on which we perform a manual scan of video devices and update our usage
	// counts in the monitor.
	DefaultReconcilerInterval = 30 * time.Second

	// DeviceWatchPath is the directory we scan and watch for video devices.
	DeviceWatchPath = filepath.Join(linux.SysFSRoot, "class", "video4linux")
)

// NewMonitor creates a new video device monitor. It performs an initial scan for video devices on the system and
// records any current usage counts.
func NewMonitor() *Monitor {
	return &Monitor{
		active: make(map[string]int),
	}
}

// Run will trigger the monitor to start monitoring for webcam events. It returns a channel on which webcam events can
// be received.
func (m *Monitor) Run(ctx context.Context) chan Event {
	slogctx.FromCtx(ctx).Debug("Started webcam monitor.",
		slog.Time("started", time.Now().UTC()),
	)
	// Set up event channel.
	m.eventCh = make(chan Event)

	// Run reconciler in background.
	go m.reconciler(ctx, DefaultReconcilerInterval)

	// Run main loop.
	go func() {
		for {
			select {
			case <-ctx.Done():
				slogctx.FromCtx(ctx).Debug("Stopped webcam monitor.",
					slog.Time("stopped", time.Now().UTC()),
				)
				close(m.eventCh)
				return
			default:
				// Get current devices.
				devices, _ := GetVideoDevices()

				// Scan usage of video devices.
				counts := ScanVideoDevices(devices)

				// Record usage counts.
				for device := range slices.Values(devices) {
					if count, ok := counts[device]; ok {
						m.setCount(device, count)
					}
				}

				// Watch devices.
				if err := m.watchDevices(
					ctx,
					devices,
					DeviceWatchPath,
				); err != nil {
					slogctx.FromCtx(ctx).Warn("Watch devices error, restarting...",
						slog.Any("error", err),
					)
					time.Sleep(2 * time.Second)
				}
			}
		}
	}()

	return m.eventCh
}

func (m *Monitor) reconciler(ctx context.Context, interval time.Duration) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.Tick(interval):
			devices, err := GetVideoDevices()
			if err != nil {
				slogctx.FromCtx(ctx).Warn("Could not get video devices.",
					slog.Any("error", err),
				)
			}
			counts := ScanVideoDevices(devices)
			for device := range slices.Values(devices) {
				if count, ok := counts[device]; ok {
					m.setCount(device, count)
				}
			}
		}
	}
}

func (m *Monitor) opened(ctx context.Context, dev string) {
	m.mu.Lock()
	prev := m.active[dev]
	m.active[dev]++
	m.mu.Unlock()
	if prev == 0 {
		slogctx.FromCtx(ctx).Debug("Webcam in use.",
			slog.String("device", dev),
		)
		m.emit(StatusActive, dev, "webcam now IN USE")
	}
}

func (m *Monitor) closed(ctx context.Context, dev string) {
	m.mu.Lock()
	if m.active[dev] > 0 {
		m.active[dev]--
	}
	cur := m.active[dev]
	m.mu.Unlock()
	if cur == 0 {
		slogctx.FromCtx(ctx).Debug("Webcam idle.",
			slog.String("device", dev),
		)
		m.emit(StatusIdle, dev, "webcam now idle")
	}
}

func (m *Monitor) setCount(dev string, count int) {
	m.mu.Lock()
	prev := m.active[dev]
	m.active[dev] = count
	m.mu.Unlock()

	switch {
	case prev == 0 && count > 0:
		m.emit(StatusActive, dev, fmt.Sprintf("webcam now IN USE (%d open fd(s))", count))
	case prev > 0 && count == 0:
		m.emit(StatusIdle, dev, "webcam now idle")
	}
}

func (m *Monitor) emit(status Status, dev, msg string) {
	if m.eventCh != nil {
		m.eventCh <- Event{
			Timestamp: time.Now().UTC(),
			Device:    dev,
			Status:    status,
			Message:   msg,
		}
	}
}

// ---------------------------------------------------------------------------
// inotify watcher
// ---------------------------------------------------------------------------

// inotifyEvent mirrors the kernel struct – 16 bytes fixed + variable Name.
type inotifyEvent struct {
	Wd     int32
	Mask   uint32
	Cookie uint32
	Len    uint32
	// Name []byte follows (length == Len, NUL-padded to 4-byte boundary)
}

const inotifyEventSize = int(unsafe.Sizeof(inotifyEvent{})) // 16

func (m *Monitor) watchDevices(ctx context.Context, devices []string, devDir string) error {
	fd, err := syscall.InotifyInit1(syscall.IN_CLOEXEC | syscall.IN_NONBLOCK)
	if err != nil {
		return fmt.Errorf("inotify_init: %w", err)
	}
	defer syscall.Close(fd)

	// wd → device path
	wdMap := make(map[int]string)

	// Watch each device file directly.
	for device := range slices.Values(devices) {
		wd, err := syscall.InotifyAddWatch(fd, device, watchMask)
		if err != nil {
			slogctx.FromCtx(ctx).Warn("Cannot watch device.",
				slog.Any("error", err),
			)
			continue
		}
		wdMap[wd] = device
		slogctx.FromCtx(ctx).Debug("Watching video device.",
			slog.String("device", device))
	}

	// Also watch the /dev directory so we notice new /dev/videoN devices
	// appearing (e.g. USB camera plugged in).
	devWd, err := syscall.InotifyAddWatch(fd, devDir,
		inotifyCreate|inotifyDelete|inotifyMovedTo|inotifyMovedFrom)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Cannot watch directory.",
			slog.Any("error", err),
		)
		devWd = -1
	}
	slogctx.FromCtx(ctx).Debug("Watching for hotplug events.",
		slog.String("directory", devDir),
	)

	buf, ok := bufPool.Get().([]byte)
	if !ok {
		return fmt.Errorf("unable to allocate buffer: %w", err)
	}
	defer func() {
		for i := range buf {
			buf[i] = 0
		}
		bufPool.Put(buf)
	}()

	// Use epoll for efficient blocking
	epfd, err := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	if err != nil {
		return fmt.Errorf("epoll_create: %w", err)
	}
	defer syscall.Close(epfd)

	ev := syscall.EpollEvent{Events: syscall.EPOLLIN, Fd: int32(fd)}
	if err := syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, fd, &ev); err != nil {
		return fmt.Errorf("epoll_ctl: %w", err)
	}

	events := make([]syscall.EpollEvent, 8)

	for {
		n, err := syscall.EpollWait(epfd, events, 5000) // 5 s timeout
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			return fmt.Errorf("epoll_wait: %w", err)
		}
		if n == 0 {
			// Timeout – periodically reconcile with /proc in case we missed an event
			devList, _ := GetVideoDevices()
			counts := ScanVideoDevices(devList)
			for _, dev := range devList {
				m.setCount(dev, counts[dev])
			}
			continue
		}

		for {
			nr, err := syscall.Read(fd, buf)
			if err != nil {
				if err == syscall.EAGAIN || err == syscall.EWOULDBLOCK {
					break
				}
				return fmt.Errorf("read inotify fd: %w", err)
			}
			if nr == 0 {
				break
			}

			offset := 0
			for offset+inotifyEventSize <= nr {
				raw := (*inotifyEvent)(unsafe.Pointer(&buf[offset]))
				nameLen := int(raw.Len)

				var name string
				if nameLen > 0 && offset+inotifyEventSize+nameLen <= nr {
					nameBuf := buf[offset+inotifyEventSize : offset+inotifyEventSize+nameLen]
					// NUL-terminate
					nul := 0
					for nul < len(nameBuf) && nameBuf[nul] != 0 {
						nul++
					}
					name = string(nameBuf[:nul])
				}

				offset += inotifyEventSize + nameLen

				mask := raw.Mask
				wd := raw.Wd

				if int(wd) == devWd {
					// Event on /dev – new or removed video device
					if !strings.HasPrefix(name, "video") {
						continue
					}
					devPath := filepath.Join(devDir, name)
					if mask&(inotifyCreate|inotifyMovedTo) != 0 {
						slogctx.FromCtx(ctx).Debug("New device.",
							slog.String("device", devPath),
						)
						newWd, err := syscall.InotifyAddWatch(fd, devPath, watchMask)
						if err == nil {
							wdMap[newWd] = devPath
						}
					} else if mask&(inotifyDelete|inotifyMovedFrom) != 0 {
						slogctx.FromCtx(ctx).Debug("Device removed.",
							slog.String("device", devPath),
						)
						m.setCount(devPath, 0)
					}
					continue
				}

				dev, ok := wdMap[int(wd)]
				if !ok {
					continue
				}

				switch {
				case mask&inotifyOpen != 0:
					m.opened(ctx, dev)
				case mask&(inotifyCloseWrite|inotifyCloseNoWrt) != 0:
					m.closed(ctx, dev)
				case mask&inotifyDeleteSelf != 0:
					m.setCount(dev, 0)
				}
			}
			_ = binary.LittleEndian // keep import used
		}
	}
}

// GetVideoDevices returns the set of /dev/video* paths on the system.
func GetVideoDevices() ([]string, error) {
	matches, err := filepath.Glob(DeviceWatchPath + "/*")
	if err != nil {
		return nil, fmt.Errorf("get video device matches: %w", err)
	}
	// Filter to only character devices
	var out []string
	for match := range slices.Values(matches) {
		fi, err := os.Stat(match)
		if err != nil {
			continue
		}
		if fi.Mode()&os.ModeCharDevice != 0 {
			out = append(out, match)
		}
	}
	return out, nil
}

// ScanVideoDevices returns a map of device path → open fd count based on /proc.
// It can only read fds for processes owned by the current uid (no root needed).
func ScanVideoDevices(devices []string) map[string]int {
	// Build rdev → path lookup
	rdevMap := make(map[uint64]string)
	for device := range slices.Values(devices) {
		if rdev, err := devNumForPath(device); err == nil {
			rdevMap[rdev] = device
		}
	}

	counts := make(map[string]int)

	entries, err := os.ReadDir(linux.ProcFSRoot)
	if err != nil {
		return counts
	}

	for entry := range slices.Values(entries) {
		if !entry.IsDir() {
			continue
		}
		pid := entry.Name()
		if _, err := strconv.Atoi(pid); err != nil {
			continue // not a PID directory
		}

		fdDir := filepath.Join(linux.ProcFSRoot, pid, "fd")
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			// Permission denied for other users' processes – expected, skip
			continue
		}

		for fd := range slices.Values(fds) {
			link := filepath.Join(fdDir, fd.Name())
			target, err := os.Readlink(link)
			if err != nil {
				continue
			}
			// Resolve the rdev of the symlink target
			var st syscall.Stat_t
			if err := syscall.Stat(target, &st); err != nil {
				continue
			}
			if dev, ok := rdevMap[st.Rdev]; ok {
				counts[dev]++
			}
		}
	}
	return counts
}

// devNumForPath returns the device number (rdev) for a device file path.
func devNumForPath(p string) (uint64, error) {
	var st syscall.Stat_t
	if err := syscall.Stat(p, &st); err != nil {
		return 0, fmt.Errorf("stat device: %w", err)
	}
	return st.Rdev, nil
}

var bufPool = sync.Pool{
	New: func() any {
		return make([]byte, 4096)
	},
}
