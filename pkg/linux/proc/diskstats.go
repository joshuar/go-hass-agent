// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package diskstats

import (
	"errors"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

//go:generate stringer -type=DiskStat -output diskStatStrings.go -linecomment
const (
	TotalReads             DiskStat = iota + 3 // Total reads completed
	TotalReadsMerged                           // Total reads merged
	TotalSectorsRead                           // Total sectors read
	TotalTimeReading                           // Total milliseconds spent reading
	TotalWrites                                // Total writes completed
	TotalWritesMerged                          // Total writes merged
	TotalSectorsWritten                        // Total sectors written
	TotalTimeWriting                           // Total milliseconds spent writing
	ActiveIOs                                  // I/Os currently in progress
	ActiveIOTime                               // Milliseconds elapsed spent doing I/Os
	ActiveIOTimeWeighted                       // Milliseconds elapsed spent doing I/Os (weighted)
	TotalDiscardsCompleted                     // Total discards completed
	TotalDiscardsMerged                        // Total discards merged
	TotalSectorsDiscarded                      // Total sectors discarded
	TotalTimeDiscarding                        // Total milliseconds spent discarding
	TotalFlushRequests                         // Total flush requests completed
	TotalTimeFlushing                          // Total milliseconds spent flushing
)

var ErrDeviceNotFound = errors.New("device not found in diskstats")

type DiskStat int

// ReadDiskStats reads /proc/diskstats and returns a map of devices, which in
// turn contains a map of disk stats for the given device. If there was a
// problem reading /proc/diskstats, a non-nil error will be returned.
func ReadDiskStats() (map[string]map[DiskStat]uint64, error) {
	data, err := os.ReadFile("/proc/diskstats")
	if err != nil {
		return nil, err
	}

	stats := make(map[string]map[DiskStat]uint64)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines[:len(lines)-1] {
		fields := strings.Split(line, " ")
		fields = slices.DeleteFunc(fields, func(n string) bool {
			return n == ""
		})
		device := fields[2]
		stats[device] = make(map[DiskStat]uint64)
		for i, f := range fields {
			if i < 3 {
				continue
			}
			stat := DiskStat(i)
			readVal, err := strconv.ParseUint(f, 10, 64)
			if err != nil {
				log.Warn().
					Err(err).
					Str("stat", stat.String()).
					Str("device", device).
					Msg("Unable to read disk stat.")
			}
			stats[device][stat] = readVal
		}
	}
	return stats, nil
}

// DeviceStats will retrieve the stats for the given device. If the device is
// not found, it will return an ErrDeviceNotFound error or another error as
// appropriate for any failure.
func DeviceStats(device string) (map[DiskStat]uint64, error) {
	allStats, err := ReadDiskStats()
	if err != nil {
		return nil, err
	}
	stats, ok := allStats[device]
	if !ok {
		return nil, ErrDeviceNotFound
	}
	return stats, nil
}
