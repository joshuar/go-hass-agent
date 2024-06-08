// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package diskstats

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

//go:generate stringer -type=Stat -output diskStatStrings.go -linecomment
const (
	TotalReads             Stat = iota // Total reads completed
	TotalReadsMerged                   // Total reads merged
	TotalSectorsRead                   // Total sectors read
	TotalTimeReading                   // Total milliseconds spent reading
	TotalWrites                        // Total writes completed
	TotalWritesMerged                  // Total writes merged
	TotalSectorsWritten                // Total sectors written
	TotalTimeWriting                   // Total milliseconds spent writing
	ActiveIOs                          // I/Os currently in progress
	ActiveIOTime                       // Milliseconds elapsed spent doing I/Os
	ActiveIOTimeWeighted               // Milliseconds elapsed spent doing I/Os (weighted)
	TotalDiscardsCompleted             // Total discards completed
	TotalDiscardsMerged                // Total discards merged
	TotalSectorsDiscarded              // Total sectors discarded
	TotalTimeDiscarding                // Total milliseconds spent discarding
	TotalFlushRequests                 // Total flush requests completed
	TotalTimeFlushing                  // Total milliseconds spent flushing
)

// Stat represents a specific statistic recorded by the kernel for the
// associated disk.
type Stat int

type Device struct {
	ID        string `json:"device_id"`
	SysFSPath string `json:"sysfs_path,omitempty"`
	Model     string `json:"device_model,omitempty"`
}

var (
	ErrDeviceNotFound = errors.New("device not found in diskstats")
	ErrSplitFailed    = errors.New("could not split into lines")
)

// ReadDiskStatsFromProcFS reads /proc/diskstats and returns a map of devices,
// which in turn contains a map of disk stats for the given device. If there was
// a problem reading /proc/diskstats, a non-nil error will be returned. Note
// that /proc/diskstats contains all block devices, all virtual devices, and all
// partitons.
//
//nolint:mnd
func ReadDiskStatsFromProcFS() (map[Device]map[Stat]uint64, error) {
	data, err := os.ReadFile("/proc/diskstats")
	if err != nil {
		return nil, fmt.Errorf("unable to read /proc/disktats: %w", err)
	}

	stats := make(map[Device]map[Stat]uint64)
	lines := strings.Split(string(data), "\n")

	if lines == nil {
		return nil, ErrSplitFailed
	}

	for _, line := range lines[:len(lines)-1] {
		fields := strings.Split(line, " ")
		fields = slices.DeleteFunc(fields, func(n string) bool {
			return n == ""
		})
		device := Device{
			ID:        fields[2],
			SysFSPath: "",
			Model:     "",
		}
		stats[device] = make(map[Stat]uint64)

		for i, field := range fields {
			if i < 3 {
				continue
			}

			stat := Stat(i - 3)

			readVal, err := strconv.ParseUint(field, 10, 64)
			if err != nil {
				log.Warn().
					Err(err).
					Str("stat", stat.String()).
					Str("device", device.ID).
					Msg("Unable to read disk stat.")
			}

			stats[device][stat] = readVal
		}
	}

	return stats, nil
}

// ReadDiskStatsFromSysFS will read the individual stat files for block devices
// from sysfs and returns a map of devices, which in turn contains a map of disk
// stats for the given device. It will filter out stats for some virtual
// devices. It also returns a "total" device containing the sum for each
// statistic from all devices.
//
//nolint:cyclop
func ReadDiskStatsFromSysFS() (map[Device]map[Stat]uint64, error) {
	allDevices := make(map[Device]map[Stat]uint64)

	devices, err := filepath.Glob("/sys/block/*/stat")
	if err != nil {
		return nil, fmt.Errorf("unable to read files under /sys/block/*/stat: %w", err)
	}

	total := Device{
		ID:        "total",
		SysFSPath: "",
		Model:     "",
	}
	totalStats := make(map[Stat]uint64)

	for _, dev := range devices {
		// Convert the device path into the id by removing the leading and
		// trailing path elements.
		id := dev
		id = strings.TrimPrefix(id, "/sys/block/")
		id = strings.TrimSuffix(id, "/stat")
		// Exclude loop and ram device statistics
		if strings.HasPrefix(id, "loop") || strings.HasPrefix(id, "ram") {
			continue
		}

		device := Device{
			ID:        id,
			SysFSPath: "/sys/block/" + id,
			Model:     "",
		}
		// Try to read the model from the appropriate file. Otherwise just leave
		// it empty.
		if model, err := os.ReadFile("/sys/block/" + id + "/device/model"); err == nil {
			device.Model = strings.TrimSpace(string(model))
		}
		// Read the stats file.
		data, err := os.ReadFile(dev)
		if err != nil {
			log.Warn().Err(err).Str("device", id).Msg("Unable to read stats for device.")

			continue
		}
		// Parse the stats file fields into the relevant stats.
		stats := make(map[Stat]uint64)
		fields := bytes.Fields(data)

		for i, field := range fields {
			stat := Stat(i)

			readVal, err := strconv.ParseUint(string(field), 10, 64)
			if err != nil {
				log.Warn().
					Err(err).
					Str("stat", stat.String()).
					Str("id", dev).
					Msg("Unable to parse device stat.")
			}

			stats[stat] = readVal
			// Don't include virtual devices in totals
			if strings.HasPrefix(id, "dm") || strings.HasPrefix(id, "md") {
				continue
			}

			totalStats[stat] += readVal
		}
		// Add this device to the device stats map.
		allDevices[device] = stats
	}

	allDevices[total] = totalStats

	return allDevices, nil
}

// DeviceStats will retrieve the stats for the given device. If the device is
// not found, it will return an ErrDeviceNotFound error or another error as
// appropriate for any failure.
func DeviceStats(device string) (map[Stat]uint64, error) {
	allStats, err := ReadDiskStatsFromProcFS()
	if err != nil {
		return nil, err
	}

	for k, v := range allStats {
		if k.ID == device {
			return v, nil
		}
	}

	return nil, ErrDeviceNotFound
}
