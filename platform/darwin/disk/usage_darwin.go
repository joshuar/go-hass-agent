package disk

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"golang.org/x/sys/unix"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	commondisk "github.com/joshuar/go-hass-agent/platform/common/disk"
	"github.com/joshuar/go-hass-agent/platform/darwin"
	"github.com/joshuar/go-hass-agent/scheduler"
)

const (
	usageUpdateInterval = time.Minute
	usageUpdateJitter   = 10 * time.Second

	usageWorkerID   = "disk_usage_sensors"
	usageWorkerDesc = "Disk usage stats"
)

// ignoredFlags are mount flags marking a volume as not worth reporting on.
// MNT_DONTBROWSE is what macOS itself sets on volumes that should not appear
// in Finder: the internal /System/Volumes/* firmlink targets, Time Machine
// backup volumes, and the .timemachine snapshot mounts. MNT_QUARANTINE is
// set on mounts originating from Gatekeeper-quarantined content, which in
// practice means a disk image the user downloaded and double-clicked.
// Note MNT_SNAPSHOT is deliberately not included: the sealed system volume
// mounted at "/" is itself a snapshot.
const ignoredFlags uint32 = unix.MNT_DONTBROWSE | unix.MNT_QUARANTINE

// ignoredFilesystems are filesystem types that should never be reported as
// disk usage sensors, regardless of mountpoint: pseudo filesystems and
// network shares.
var ignoredFilesystems = []string{"devfs", "autofs", "smbfs", "nfs", "afpfs", "webdav"}

// ignoredMounts are additional mountpoints (or path prefixes) to skip. macOS
// flags its own internal volumes itself (see ignoredFlags), so this is empty
// by default and exists for users to exclude specific mounts via preferences.
var ignoredMounts []string

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

	worker := &commondisk.Worker{
		WorkerMetadata: models.SetWorkerMetadata(usageWorkerID, usageWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{
			Trigger: scheduler.NewPollTriggerWithJitter(pollInterval, usageUpdateJitter),
		},
		DataSource: darwin.DataSrcGetfsstat,
		Disabled:   prefs.Disabled,
		GetMounts: func(ctx context.Context) ([]commondisk.Mount, error) {
			return getMounts(ctx, prefs.IgnoredMounts)
		},
	}

	return worker, nil
}

func getMounts(ctx context.Context, ignoredMounts []string) ([]commondisk.Mount, error) {
	initialCount, err := unix.Getfsstat(nil, unix.MNT_NOWAIT)
	if err != nil {
		return nil, err
	}
	// create bigger buffer in case if something mounts inbetween
	buf := make([]unix.Statfs_t, initialCount+4)
	newCount, err := unix.Getfsstat(buf, unix.MNT_NOWAIT)
	if err != nil {
		return nil, err
	}
	if newCount == len(buf) {
		// may have been truncated, retry
		return getMounts(ctx, ignoredMounts)
	}
	result := make([]commondisk.Mount, 0, newCount)

	for _, stat := range buf[:newCount] {
		if stat.Flags&ignoredFlags != 0 {
			continue
		}

		mountpoint := unix.ByteSliceToString(stat.Mntonname[:])
		filesystem := unix.ByteSliceToString(stat.Fstypename[:])

		if slices.Contains(ignoredFilesystems, filesystem) {
			continue
		}

		if slices.ContainsFunc(ignoredMounts, func(blocked string) bool {
			return mountpoint == blocked || strings.HasPrefix(mountpoint, blocked)
		}) {
			continue
		}

		newMount := commondisk.Mount{
			Mountpoint: mountpoint,
			Attributes: make(map[string]any),
		}
		newMount.Attributes[commondisk.AttrDevice] = unix.ByteSliceToString(stat.Mntfromname[:])
		newMount.Attributes[commondisk.AttrFs] = filesystem
		newMount.Attributes[commondisk.AttrBlockSize] = uint64(stat.Bsize)
		newMount.Attributes[commondisk.AttrBlocksTotal] = stat.Blocks
		newMount.Attributes[commondisk.AttrBytesTotal] = stat.Blocks * uint64(stat.Bsize)
		newMount.Attributes[commondisk.AttrBlocksFree] = stat.Bfree
		newMount.Attributes[commondisk.AttrBytesFree] = stat.Bfree * uint64(stat.Bsize)
		newMount.Attributes[commondisk.AttrBlocksAvail] = stat.Bavail
		newMount.Attributes[commondisk.AttrInodesTotal] = stat.Files
		newMount.Attributes[commondisk.AttrInodesFree] = stat.Ffree

		result = append(result, newMount)
	}
	return result, nil
}
