// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"

	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/pkg/linux/hwmon"
)

func main() {
	cpu, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create CPU profile.")
	}
	if err := pprof.StartCPUProfile(cpu); err != nil {
		log.Fatal().Err(err).Msg("Could not start CPU profiling.")
	}
	trc, err := os.Create("trace.prof")
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create trace profile.")
	}
	if err = trace.Start(trc); err != nil {
		log.Fatal().Err(err).Msg("Could not start trace profiling.")
	}
	for _, s := range hwmon.GetAllSensors() {
		println(s.String())
	}

	pprof.StopCPUProfile()
	trace.Stop()

	heap, err := os.Create("heap.prof")
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create heap profile.")
	}

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	printMemStats(&ms)

	if err := pprof.WriteHeapProfile(heap); err != nil {
		log.Fatal().Err(err).Msg("Cannot write heap profile.")
	}
	_ = heap.Close()
}

func printMemStats(ms *runtime.MemStats) {
	log.Info().Msgf("Mem stats: alloc=%s total_alloc=%s sys=%s "+
		"heap_alloc=%s heap_sys=%s heap_idle=%s heap_released=%s heap_in_use=%s "+
		"stack_in_use=%s stack_sys=%s "+
		"mspan_sys=%s mcache_sys=%s buck_hash_sys=%s gc_sys=%s other_sys=%s "+
		"mallocs_n=%d frees_n=%d heap_objects_n=%d gc_cpu_fraction=%.2f",
		formatMemory(ms.Alloc), formatMemory(ms.TotalAlloc), formatMemory(ms.Sys),
		formatMemory(ms.HeapAlloc), formatMemory(ms.HeapSys),
		formatMemory(ms.HeapIdle), formatMemory(ms.HeapReleased), formatMemory(ms.HeapInuse),
		formatMemory(ms.StackInuse), formatMemory(ms.StackSys),
		formatMemory(ms.MSpanSys), formatMemory(ms.MCacheSys), formatMemory(ms.BuckHashSys),
		formatMemory(ms.GCSys), formatMemory(ms.OtherSys),
		ms.Mallocs, ms.Frees, ms.HeapObjects, ms.GCCPUFraction)
}

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
