// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
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
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/reugn/go-quartz/quartz"
	slogctx "github.com/veqryn/slog-context"
	"golang.org/x/sys/unix"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/class"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/platform/linux"
	"github.com/joshuar/go-hass-agent/scheduler"
)

const (
	usageUpdateInterval = time.Minute
	usageUpdateJitter   = 10 * time.Second

	usageWorkerID   = "disk_usage_sensors"
	usageWorkerDesc = "Disk usage stats"
)

const (
	diskUsageSensorIcon  = "mdi:harddisk"
	diskUsageSensorUnits = "%"
)

const (
	mountAttrDevice      = "device"
	mountAttrFs          = "filesystem_type"
	mountAttrOpts        = "mount_options"
	mountAttBlockSize    = "block_size"
	mountAttrBlocksTotal = "blocks_total"
	mountAttrBlocksFree  = "blocks_free"
	mountAttrBlocksAvail = "blocks_available"
	mountAttrInodesTotal = "inodes_total"
	mountAttrInodesFree  = "inodes_free"
)

type mount struct {
	attributes map[string]any
	mountpoint string
}

var (
	validVirtualFs       = []string{"tmpfs", "ramfs", "cifs", "smb", "nfs"}
	defaultIgnoredMounts = []string{"/tmp/crun", "/run", "/var/lib/containers", "/sys", "/proc"}
)

var (
	_ quartz.Job                  = (*usageWorker)(nil)
	_ workers.PollingEntityWorker = (*usageWorker)(nil)
)

type usageWorker struct {
	*models.WorkerMetadata
	*workers.PollingEntityWorkerData
	prefs *usageWorkerPrefs
}

type usageWorkerPrefs struct {
	WorkerPrefs

	IgnoredMounts []string `toml:"ignored_mounts"`
}

// NewUsageWorker creates a new polling sensor worker to monitor disk mount usage.
func NewUsageWorker(ctx context.Context) (workers.EntityWorker, error) {
	usageWorker := &usageWorker{
		WorkerMetadata:          models.SetWorkerMetadata(usageWorkerID, usageWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
	}

	defaultPrefs := &usageWorkerPrefs{
		IgnoredMounts: defaultIgnoredMounts,
	}
	defaultPrefs.UpdateInterval = usageUpdateInterval.String()

	var err error
	usageWorker.prefs, err = workers.LoadWorkerPreferences(usageWorkerPreferencesID, defaultPrefs)
	if err != nil {
		return nil, fmt.Errorf("could not load disk usage worker preferences: %w", err)
	}

	pollInterval, err := time.ParseDuration(usageWorker.prefs.UpdateInterval)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Invalid polling interval, using default",
			slog.String("worker", usageWorkerID),
			slog.String("given_interval", usageWorker.prefs.UpdateInterval),
			slog.String("default_interval", usageUpdateInterval.String()))

		pollInterval = usageUpdateInterval
	}
	usageWorker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, usageUpdateJitter)

	return usageWorker, nil
}

func (w *usageWorker) Execute(ctx context.Context) error {
	mounts, err := getMounts(ctx, w.prefs.IgnoredMounts)
	if err != nil {
		return fmt.Errorf("could not get mount points: %w", err)
	}

	for mount := range slices.Values(mounts) {
		usedBlocks := mount.attributes[mountAttrBlocksTotal].(uint64) - mount.attributes[mountAttrBlocksFree].(uint64) //nolint:lll,forcetypeassert
		usedPc := float64(usedBlocks) / float64(mount.attributes[mountAttrBlocksTotal].(uint64)) * 100                 //nolint:forcetypeassert

		if math.IsNaN(usedPc) {
			continue
		}
		w.OutCh <- newDiskUsageSensor(ctx, mount, usedPc)
	}
	return nil
}

func (w *usageWorker) IsDisabled() bool {
	return w.prefs.Disabled
}

func (w *usageWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start disk usage worker: %w", err)
	}
	return w.OutCh, nil
}

func newDiskUsageSensor(ctx context.Context, mount *mount, value float64) models.Entity {
	mount.attributes["data_source"] = linux.DataSrcProcFS

	usedBlocks := mount.attributes[mountAttrBlocksTotal].(uint64) - mount.attributes[mountAttrBlocksFree].(uint64) //nolint:lll,forcetypeassert
	mount.attributes["blocks_used"] = usedBlocks

	var id string

	if mount.mountpoint == "/" {
		id = "mountpoint_root"
	} else {
		id = "mountpoint" + strings.ReplaceAll(mount.mountpoint, "/", "_")
	}

	return sensor.NewSensor(ctx,
		sensor.WithName("Mountpoint "+mount.mountpoint+" Usage"),
		sensor.WithID(id),
		sensor.WithUnits(diskUsageSensorUnits),
		sensor.WithStateClass(class.StateTotal),
		sensor.WithIcon(diskUsageSensorIcon),
		sensor.WithState(math.Round(value/0.05)*0.05),
		sensor.WithAttributes(mount.attributes),
	)
}

func (m *mount) getMountInfo() error {
	var stats unix.Statfs_t

	err := unix.Statfs(m.mountpoint, &stats)
	if err != nil {
		return fmt.Errorf("getMountInfo: %w", err)
	}

	m.attributes[mountAttBlockSize] = stats.Bsize
	m.attributes[mountAttrBlocksTotal] = stats.Blocks
	m.attributes[mountAttrBlocksFree] = stats.Bfree
	m.attributes[mountAttrBlocksAvail] = stats.Bavail
	m.attributes[mountAttrInodesTotal] = stats.Files
	m.attributes[mountAttrInodesFree] = stats.Ffree

	return nil
}

func getFilesystems() ([]string, error) {
	data, err := os.Open(filepath.Join(linux.ProcFSRoot, "filesystems"))
	if err != nil {
		return nil, fmt.Errorf("getFilesystems: %w", err)
	}

	defer data.Close() //nolint:errcheck

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

	return filesystems, nil
}

func getMounts(ctx context.Context, ignoredMounts []string) ([]*mount, error) {
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

	var mounts []*mount
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
		// If the fs is in our valid filesystems, It should be in the list of
		// valid filesystems and not one of the blocked mountpoints.
		if slices.Contains(filesystems, filesystem) &&
			!slices.ContainsFunc(ignoredMounts, func(blockedMount string) bool { return strings.HasPrefix(mountpoint, blockedMount) }) {
			validmount := &mount{
				mountpoint: mountpoint,
				attributes: make(map[string]any),
			}
			validmount.attributes[mountAttrDevice] = device
			validmount.attributes[mountAttrFs] = filesystem
			validmount.attributes[mountAttrOpts] = opts

			if err := validmount.getMountInfo(); err != nil {
				slogctx.FromCtx(ctx).
					With(slog.String("worker", usageWorkerID)).
					Debug("Error getting mount info.", slog.Any("error", err))
			} else {
				mounts = append(mounts, validmount)
			}
		}
	}

	if err := data.Close(); err != nil {
		slogctx.FromCtx(ctx).
			With(slog.String("worker", usageWorkerID)).
			Debug("Failed to close mounts file.", slog.Any("error", err))
	}

	return mounts, nil
}
