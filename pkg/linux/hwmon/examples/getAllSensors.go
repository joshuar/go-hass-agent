// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//nolint:all
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"

	"github.com/joshuar/go-hass-agent/pkg/linux/hwmon"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	cpu, err := os.Create("cpu.prof")
	if err != nil {
		slog.Warn("Cannot create CPU profile.", "error", err.Error())
	}
	if err := pprof.StartCPUProfile(cpu); err != nil {
		slog.Warn("Could not start CPU profiling.", "error", err.Error())
	}
	trc, err := os.Create("trace.prof")
	if err != nil {
		slog.Warn("Cannot create trace profile.", "error", err.Error())
	}
	if err = trace.Start(trc); err != nil {
		slog.Warn("Could not start trace profiling.", "error", err.Error())
	}
	sensors, err := hwmon.GetAllSensors(context.Background())
	if err != nil && len(sensors) > 0 {
		slog.Warn("Errors fetching some chip/sensor values.", "error", err.Error())
	}
	if err != nil && len(sensors) == 0 {
		slog.Error("Could not retrieve any chip/sensor values.", "error", err.Error())
		os.Exit(-1)
	}
	for _, s := range sensors {
		println(s.String())
	}

	pprof.StopCPUProfile()
	trace.Stop()

	heap, err := os.Create("heap.prof")
	if err != nil {
		slog.Warn("Cannot create heap profile.", "error", err.Error())
	}

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	// printMemStats(&ms)

	if err := pprof.WriteHeapProfile(heap); err != nil {
		slog.Warn("Cannot write heap profile.", "error", err.Error())
	}
	_ = heap.Close()
}

// func printMemStats(ms *runtime.MemStats) {
// 	log.Info().Msgf("Mem stats: alloc=%s total_alloc=%s sys=%s "+
// 		"heap_alloc=%s heap_sys=%s heap_idle=%s heap_released=%s heap_in_use=%s "+
// 		"stack_in_use=%s stack_sys=%s "+
// 		"mspan_sys=%s mcache_sys=%s buck_hash_sys=%s gc_sys=%s other_sys=%s "+
// 		"mallocs_n=%d frees_n=%d heap_objects_n=%d gc_cpu_fraction=%.2f",
// 		formatMemory(ms.Alloc), formatMemory(ms.TotalAlloc), formatMemory(ms.Sys),
// 		formatMemory(ms.HeapAlloc), formatMemory(ms.HeapSys),
// 		formatMemory(ms.HeapIdle), formatMemory(ms.HeapReleased), formatMemory(ms.HeapInuse),
// 		formatMemory(ms.StackInuse), formatMemory(ms.StackSys),
// 		formatMemory(ms.MSpanSys), formatMemory(ms.MCacheSys), formatMemory(ms.BuckHashSys),
// 		formatMemory(ms.GCSys), formatMemory(ms.OtherSys),
// 		ms.Mallocs, ms.Frees, ms.HeapObjects, ms.GCCPUFraction)
// }

//nolint:varnamelen,wsl,nlreturn
//revive:disable:unexported-naming
func formatMemory(memBytes uint64) string {
	const Kb = 1024
	const Mb = Kb * 1024

	if memBytes < Kb {
		return fmt.Sprintf("%db", memBytes)
	}
	if memBytes < Mb {
		return fmt.Sprintf("%dkb", memBytes/Kb)
	}
	return fmt.Sprintf("%dmb", memBytes/Mb)
}
