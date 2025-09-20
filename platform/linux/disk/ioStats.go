// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//go:generate go tool golang.org/x/tools/cmd/stringer -type=stat -output ioStats_generated.go -linecomment
package disk

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

var (
	ErrParseDevices = errors.New("could not parse devices")
	deviceMajNo     = []string{"8", "252", "253", "259"}
)

const (
	TotalReads             stat = iota // Total reads completed
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

// stat represents a specific statistic recorded by the kernel for the
// associated disk.
type stat int

type device struct {
	id        string
	sysFSPath string
	model     string
}

func (d *device) generateIdentifiers(sensorType ioSensor) (string, string) {
	r := []rune(d.id)
	name := string(append([]rune{unicode.ToUpper(r[0])}, r[1:]...)) + " " + sensorType.String()
	id := strings.ToLower(d.id + "_" + strings.ReplaceAll(sensorType.String(), " ", "_"))

	return name, id
}

func (d *device) generateAttributes() models.Attributes {
	attributes := models.Attributes{
		"data_source": linux.DataSrcSysFS,
	}

	// Add attributes from device if available.
	if d.model != "" {
		attributes["device_model"] = d.model
	}

	if d.sysFSPath != "" {
		attributes["sysfs_path"] = d.sysFSPath
	}

	return attributes
}

func getDeviceNames() ([]string, error) {
	data, err := os.Open(filepath.Join(linux.ProcFSRoot, "partitions"))
	if err != nil {
		return nil, fmt.Errorf("getDevices: %w", err)
	}

	defer data.Close() //nolint:errcheck

	var devices []string

	partitions := bufio.NewScanner(data)
	// Skip first two lines (header + blank line).
	for range 2 {
		partitions.Scan()
	}
	// Read remaining lines.
	for partitions.Scan() {
		line := bufio.NewScanner(bytes.NewReader(partitions.Bytes()))
		line.Split(bufio.ScanWords)

		var cols []string

		for line.Scan() {
			cols = append(cols, line.Text())
		}

		if len(cols) == 0 {
			return devices, ErrParseDevices
		}

		if validDeviceNo(cols) {
			devices = append(devices, cols[3])
		}
	}

	return devices, nil
}

func validDeviceNo(details []string) bool {
	if slices.Contains(deviceMajNo, details[0]) {
		if details[1] == "0" {
			return true
		}
	}

	return false
}

func getDevice(deviceName string) (*device, map[stat]uint64, error) {
	// Create a new device.
	dev := &device{
		id:        deviceName,
		sysFSPath: filepath.Join(linux.SysFSRoot, "block", deviceName),
	}

	// Try to read the model from the appropriate file. Otherwise just leave
	// it empty.
	if model, err := os.ReadFile(dev.sysFSPath + "/device/model"); err == nil {
		dev.model = strings.TrimSpace(string(model))
	}

	data, err := os.ReadFile(dev.sysFSPath + "/stat")
	if err != nil {
		return dev, nil, fmt.Errorf("getDeviceStats: %w", err)
	}

	line := bufio.NewScanner(bytes.NewReader(data))
	line.Split(bufio.ScanWords)

	stats := make(map[stat]uint64)
	statno := stat(0)
	// Parse the rest as stats.
	for line.Scan() {
		readVal, err := strconv.ParseUint(line.Text(), 10, 64)
		if err != nil {
			slog.Warn("Unable to parse device stat.",
				slog.String("device", dev.id),
				slog.String("stat", line.Text()),
				slog.Any("error", err))
		} else {
			stats[statno] = readVal
		}

		statno++
	}

	return dev, stats, nil
}
