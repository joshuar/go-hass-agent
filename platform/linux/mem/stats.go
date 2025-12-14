// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//go:generate go tool stringer -type=memStatID -output stats.gen.go -linecomment
package mem

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/platform/linux"
)

// All of the possible memory statistics. We map these to an iota which makes it
// easier to refer to them elsewhere in code. The iota is also used to generate
// a name for the statistic, for the associated sensor name.
const (
	memTotal          memStatID = iota //          Memory Total
	memFree                            //          Memory Free
	memAvailable                       //      Memory Available
	memBuffered                        //           Memory Buffered
	memCached                          //            Memory Cached
	swapCached                         //        Swap Cached
	memActive                          //            Memory Active
	memInactive                        //          Memory Inactive
	memAnonActive                      //      Anonymous Memory Active
	memAnonInactive                    //    Anonymous Memory Inactive
	memFileActive                      //      File Active Memory
	memFileInactive                    //    File Inactive Memory
	memUnevictable                     //       Unevictable Memory
	memLocked                          //           Locked Memory
	swapTotal                          //         Swap Total
	swapFree                           //          Swap Free
	zswapTotal                         //             Zswap Total
	zswapUsed                          //          Zswapped Used
	memDirty                           //             Dirty Memory
	memWriteback                       //         Writeback Memory
	memAnonPages                       //         Anonymous Page Tables Memory
	memMapped                          //            mmap Memory
	memShmem                           //             shmem Memory
	memKReclaimable                    //      Kernel Memory Reclaimable
	memSlab                            //              Kernel Slab Memory
	memSReclaimable                    //      Kernel Slab Memory Reclaimable
	memSUnreclaim                      //        Kernel Slab Memory Unreclaimable
	memKernelStack                     //       Kernel Stack Memory
	memPageTables                      //        Page Tables Memory
	memSecPageTables                   //     Secure Page Tables Memory
	memNFSUnstable                     //      NFS Pages Memory
	memBounce                          //            Block Device Bounce Buffer Memory
	memWritebackTmp                    //      FUSE Temporary Writeback Buffer Memory
	memCommitLimit                     //       Commit Limit Total
	memCommittedAS                     //      Commit Limit Allocated
	vmallocTotal                       //      Vmalloc Total Memory
	vmallocUsed                        //       Vmalloc Used Memory
	vmallocChunk                       //      Vmalloc Largest Unused Chunk
	memPercpu                          //            Percpu Memory
	memCorrupted                       // Memory Corrupted
	memAnonHugePages                   //     Anonymouse Huge Pages Memory
	memShmemHugePages                  //    shmem Huge Pages Memory
	memShmemPmdMapped                  //    shmem User Space Huge Pages Memory
	memFileHugePages                   //   File Huge Pages Memory
	memFilePmdMapped                   //     File User Space Huge Pages Memory
	memCmaTotal                        //          Contiguous Memory Allocator Pages Total
	memCmaFree                         //           Contiguous Memory Allocator Pages Free
	memUnaccepted                      //        Unaccepted Memory
	memHugePagesTotal                  //   Huge Pages Total
	memHugePagesFree                   //     Huge Pages Free
	memHugePagesRsvd                   //    Huge Pages Reserved
	memHugePagesSurp                   //    Huge Pages Surplus
	memHugepagesize                    //      Huge Page Size
	memHugetlb                         //           Huge Page TLB
	memDirectMap4k                     //       Kernel 4kB Pages
	memDirectMap2M                     //       Kernel 2MB Pages
	memDirectMap1G                     //       Kernel 1GB Pages
)

type memStatID int

// statNames maps the names/id in the statistics file to our internal memStatID.
var statNames = map[string]memStatID{
	"MemTotal":          memTotal,
	"MemFree":           memFree,
	"MemAvailable":      memAvailable,
	"Buffers":           memBuffered,
	"Cached":            memCached,
	"SwapCached":        swapCached,
	"Active":            memActive,
	"Inactive":          memInactive,
	"Active(anon)":      memAnonActive,
	"Inactive(anon)":    memAnonInactive,
	"Active(file)":      memFileActive,
	"Inactive(file)":    memFileInactive,
	"Unevictable":       memUnevictable,
	"Mlocked":           memLocked,
	"SwapTotal":         swapTotal,
	"SwapFree":          swapFree,
	"Zswap":             zswapTotal,
	"Zswapped":          zswapUsed,
	"Dirty":             memDirty,
	"Writeback":         memWriteback,
	"AnonPages":         memAnonPages,
	"Mapped":            memMapped,
	"Shmem":             memShmem,
	"KReclaimable":      memKReclaimable,
	"Slab":              memSlab,
	"SReclaimable":      memSReclaimable,
	"SUnreclaim":        memSUnreclaim,
	"KernelStack":       memKernelStack,
	"PageTables":        memPageTables,
	"SecPageTables":     memSecPageTables,
	"NFS_Unstable":      memNFSUnstable,
	"Bounce":            memBounce,
	"WritebackTmp":      memWritebackTmp,
	"CommitLimit":       memCommitLimit,
	"Committed_AS":      memCommittedAS,
	"VmallocTotal":      vmallocTotal,
	"VmallocUsed":       vmallocUsed,
	"VmallocChunk":      vmallocChunk,
	"Percpu":            memPercpu,
	"HardwareCorrupted": memCorrupted,
	"AnonHugePages":     memAnonHugePages,
	"ShmemHugePages":    memShmemHugePages,
	"ShmemPmdMapped":    memShmemPmdMapped,
	"FileHugePages":     memFileHugePages,
	"FilePmdMapped":     memFilePmdMapped,
	"CmaTotal":          memCmaTotal,
	"CmaFree":           memCmaFree,
	"Unaccepted":        memUnaccepted,
	"HugePages_Total":   memHugePagesTotal,
	"HugePages_Free":    memHugePagesFree,
	"HugePages_Rsvd":    memHugePagesRsvd,
	"HugePages_Surp":    memHugePagesSurp,
	"Hugepagesize":      memHugepagesize,
	"Hugetlb":           memHugetlb,
	"DirectMap4k":       memDirectMap4k,
	"DirectMap2M":       memDirectMap2M,
	"DirectMap1G":       memDirectMap4k,
}

var memStatFile = filepath.Join(linux.ProcFSRoot, "meminfo")

// memStat holds the value and any units for a memory statistic.
type memStat struct {
	units string
	value uint64
}

// memoryStats is a map of all the memory statistics available on this device.
type memoryStats map[memStatID]*memStat

func (m memoryStats) get(id memStatID) (uint64, string) {
	if stat, ok := m[id]; ok {
		return stat.value, stat.units
	}

	return 0, ""
}

// getMemStats will create a memoryStats map for this device.
func getMemStats(ctx context.Context) (memoryStats, error) {
	statsFH, err := os.Open(memStatFile)
	if err != nil {
		return nil, fmt.Errorf("getMemStats: %w", err)
	}

	defer statsFH.Close()

	stats := make(memoryStats)

	statsFile := bufio.NewScanner(statsFH)
	for statsFile.Scan() {
		var (
			name  string
			id    memStatID
			ok    bool
			value uint64
			err   error
		)
		// Set up word scanner for line.
		line := bufio.NewScanner(bytes.NewReader(statsFile.Bytes()))
		line.Split(bufio.ScanWords)
		// Scan the first field as the stat name.
		line.Scan()
		name = strings.Trim(line.Text(), ":")

		if id, ok = statNames[name]; !ok {
			slogctx.FromCtx(ctx).Debug("Unknown memory stat. Ignoring.", slog.String("stat", name))

			continue
		}

		// Scan the next field as the stat value.
		line.Scan()

		value, err = strconv.ParseUint(line.Text(), 10, 64)
		if err != nil {
			slogctx.FromCtx(ctx).Debug("Could not parse memory stat value.",
				slog.String("stat", name),
				slog.Any("error", err))
		}

		// If there is a third field, it will be the stat units.
		// if line.Scan() {
		// 	units = line.Text()
		// }

		stats[id] = &memStat{
			value: value * 1000, //nolint:mnd // scale to bytes for backwards compatibility
			units: memoryUsageSensorUnits,
		}
	}

	return stats, nil
}
