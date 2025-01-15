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
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/sys/unix"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/linux"
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
	validVirtualFs = []string{"tmpfs", "ramfs", "cifs", "smb", "nfs"}
	mountBlocklist = []string{"/tmp/crun", "/run"}
)

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

	return filesystems, nil
}

func getMounts(ctx context.Context) ([]*mount, error) {
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
			!slices.ContainsFunc(mountBlocklist, func(blockedMount string) bool { return strings.HasPrefix(mountpoint, blockedMount) }) {
			validmount := &mount{
				mountpoint: mountpoint,
				attributes: make(map[string]any),
			}
			validmount.attributes[mountAttrDevice] = device
			validmount.attributes[mountAttrFs] = filesystem
			validmount.attributes[mountAttrOpts] = opts

			if err := validmount.getMountInfo(); err != nil {
				logging.FromContext(ctx).
					With(slog.String("worker", usageWorkerID)).
					Debug("Error getting mount info.", slog.Any("error", err))
			} else {
				mounts = append(mounts, validmount)
			}
		}
	}

	if err := data.Close(); err != nil {
		logging.FromContext(ctx).
			With(slog.String("worker", usageWorkerID)).
			Debug("Failed to close mounts file.", slog.Any("error", err))
	}

	return mounts, nil
}
