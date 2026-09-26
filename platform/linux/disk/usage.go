// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package disk

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	slogctx "github.com/veqryn/slog-context"
	"golang.org/x/sys/unix"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/platform/common/disk"
	"github.com/joshuar/go-hass-agent/platform/linux"
	"github.com/joshuar/go-hass-agent/scheduler"
)

const (
	usageUpdateInterval = time.Minute
	usageUpdateJitter   = 10 * time.Second

	usageWorkerID   = "disk_usage_sensors"
	usageWorkerDesc = "Disk usage stats"
)

var (
	validVirtualFs = []string{"tmpfs", "ramfs", "cifs", "smb", "nfs"}
	ignoredMounts  = []string{"/tmp/crun", "/run", "/var/lib/containers", "/sys", "/proc", "/etc", "/host"}
)

type usageWorkerPrefs struct {
	WorkerPrefs `toml:",squash"`

	IgnoredMounts []string `toml:"ignored_mounts"`
}

// NewUsageWorker creates a new polling sensor worker to monitor disk mount usage.
func NewUsageWorker(_ context.Context) (workers.EntityWorker, error) {
	defaultPrefs := &usageWorkerPrefs{
		IgnoredMounts: ignoredMounts,
	}
	defaultPrefs.UpdateInterval = usageUpdateInterval.String()

	prefs, err := workers.LoadWorkerPreferences(usageWorkerPreferencesID, defaultPrefs)
	if err != nil {
		return nil, fmt.Errorf("could not load disk usage worker preferences: %w", err)
	}

	pollInterval, err := time.ParseDuration(prefs.UpdateInterval)
	if err != nil {
		pollInterval = usageUpdateInterval
	}

	worker := &disk.Worker{
		WorkerMetadata: models.SetWorkerMetadata(usageWorkerID, usageWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{
			Trigger: scheduler.NewPollTriggerWithJitter(pollInterval, usageUpdateJitter),
		},
		DataSource: linux.DataSrcProcFS,
		Disabled:   prefs.Disabled,
		GetMounts: func(ctx context.Context) ([]disk.Mount, error) {
			return getMounts(ctx, prefs.IgnoredMounts)
		},
	}

	return worker, nil
}

func getMountInfo(m *disk.Mount) error {
	var stats unix.Statfs_t

	if err := unix.Statfs(m.Mountpoint, &stats); err != nil {
		return fmt.Errorf("getMountInfo: %w", err)
	}

	m.Attributes[disk.AttrBlockSize] = uint64(stats.Bsize)
	m.Attributes[disk.AttrBlocksTotal] = stats.Blocks
	m.Attributes[disk.AttrBytesTotal] = stats.Blocks * uint64(stats.Bsize)
	m.Attributes[disk.AttrBlocksFree] = stats.Bfree
	m.Attributes[disk.AttrBytesFree] = stats.Bfree * uint64(stats.Bsize)
	m.Attributes[disk.AttrBlocksAvail] = stats.Bavail
	m.Attributes[disk.AttrInodesTotal] = stats.Files
	m.Attributes[disk.AttrInodesFree] = stats.Ffree

	return nil
}

func getFilesystems() ([]string, error) {
	data, err := os.Open(filepath.Join(linux.ProcFSRoot, "filesystems"))
	if err != nil {
		return nil, fmt.Errorf("getFilesystems: %w", err)
	}
	defer data.Close()

	var filesystems []string

	// Scan each line.
	entry := bufio.NewScanner(data)
	for entry.Scan() {
		line := bufio.NewScanner(bytes.NewReader(entry.Bytes()))
		line.Split(bufio.ScanWords)
		// Scan fields of line.
		for line.Scan() {
			switch value := line.Text(); value {
			case "nodev": // Is virtual filesystem. Check second field for fs.
				line.Scan()
				// If one of validVirtualFs, add it to tracked filesystems.
				if slices.Contains(validVirtualFs, line.Text()) {
					filesystems = append(filesystems, line.Text())
				}
			default: // Is block/regular fs. Add to tracked filesystems.
				filesystems = append(filesystems, value)
			}
		}
	}
	if err := entry.Err(); err != nil {
		return filesystems, fmt.Errorf("scan filesystems: %w", err)
	}

	return filesystems, nil
}

func getMounts(ctx context.Context, ignoredMounts []string) ([]disk.Mount, error) {
	// Get valid filesystems.
	filesystems, err := getFilesystems()
	if err != nil {
		return nil, fmt.Errorf("getMounts: %w", err)
	}

	// Open mounts file.
	data, err := os.Open(filepath.Join(linux.ProcFSRoot, "mounts"))
	if err != nil {
		return nil, fmt.Errorf("getMounts: %w", err)
	}
	defer data.Close()

	var mounts []disk.Mount
	// Scan the file.
	entry := bufio.NewScanner(data)
	for entry.Scan() {
		// Scan the line and extract first four fields device, mount, fs and
		// opts respectively..
		line := bufio.NewScanner(bytes.NewReader(entry.Bytes()))
		line.Split(bufio.ScanWords)
		line.Scan()
		device := line.Text()
		line.Scan()
		mountpoint := line.Text()
		line.Scan()
		filesystem := line.Text()
		line.Scan()
		opts := line.Text()

		// Only consider mounts with a filesystem in our list of valid filesystems.
		if slices.Contains(filesystems, filesystem) {
			// Ignore where the mountpoint is in our ignored mounts or starts with an ignored mounts string.
			if slices.ContainsFunc(ignoredMounts, func(blockedMount string) bool {
				return mountpoint == blockedMount || strings.HasPrefix(mountpoint, blockedMount)
			}) {
				continue
			}

			// Create mount details.
			validmount := disk.Mount{
				Mountpoint: mountpoint,
				Attributes: make(map[string]any),
			}
			validmount.Attributes[disk.AttrDevice] = device
			validmount.Attributes[disk.AttrFs] = filesystem
			validmount.Attributes[disk.AttrOpts] = opts

			if err := getMountInfo(&validmount); err != nil {
				slogctx.FromCtx(ctx).
					With(slog.String("worker", usageWorkerID)).
					Debug("Error getting mount info.", slog.Any("error", err))
			} else {
				mounts = append(mounts, validmount)
			}
		}
	}
	if err := entry.Err(); err != nil {
		return mounts, fmt.Errorf("scan mounts: %w", err)
	}

	return mounts, nil
}
