// Copyright (c) 2026 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//go:generate go tool stringer -type=gpuMemStatID -output gpuStats.gen.go -linecomment

package mem

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/platform/linux"
)

// All of the possible memory statistics. We map these to an iota which makes it
// easier to refer to them elsewhere in code. The iota is also used to generate
// a name for the statistic, for the associated sensor name.
const (
	memVRamTotal gpuMemStatID = iota // GPU VRam Memory Total
	memGTTTotal                      // GPU GTT Memory Total
	memVRamUsed                      // GPU VRam Memory Used
	memGTTUsed                       // GPU GTT Memory Used
)

type gpuMemStatID int

// gpuMemoryStats is a map of all the (amd) gpu memory statistics available on this device.
type gpuMemoryStats map[gpuMemStatID]uint64

func (m gpuMemoryStats) get(id gpuMemStatID) uint64 {
	if stat, ok := m[id]; ok {
		return stat
	}

	return 0
}

func getGpuStatFileLocations(card string) map[gpuMemStatID]string {
	basePath := filepath.Join(linux.SysFSRoot, "class/drm", card, "device")

	return map[gpuMemStatID]string{
		memVRamTotal: filepath.Join(basePath, "mem_info_vram_total"),
		memGTTTotal:  filepath.Join(basePath, "mem_info_gtt_total"),
		memVRamUsed:  filepath.Join(basePath, "mem_info_vram_used"),
		memGTTUsed:   filepath.Join(basePath, "mem_info_gtt_used"),
	}
}

// getAmdGpuMemStats will create a gpuMemoryStats map for an AMD-based gpu.
func getAmdGpuMemStats(ctx context.Context, card string) (gpuMemoryStats, error) {
	gpuStats := make(gpuMemoryStats)

	for id, path := range getGpuStatFileLocations(card) {
		statsFileFH, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("getAmdGpuMemStats, reading %s: %w", path, err)
		}
		defer statsFileFH.Close()

		// Set up word scanner for line.
		statsFile := bufio.NewScanner(statsFileFH)
		statsFile.Scan()

		value, err := strconv.ParseUint(statsFile.Text(), 10, 64)
		if err != nil {
			slogctx.FromCtx(ctx).Debug(
				fmt.Sprintf("Could not parse %s.", id.String()),
				slog.Any("error", err),
			)
			continue
		}

		gpuStats[id] = value
	}

	return gpuStats, nil
}
