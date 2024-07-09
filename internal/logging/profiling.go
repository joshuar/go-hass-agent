// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package logging

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"strconv"
)

type ProfileFlags map[string]string

//nolint:err113
func StartProfiling(flags ProfileFlags) error {
	for flagKey, flagVal := range flags {
		switch flagKey {
		case "webui":
			if err := startWebProfiler(flagVal); err != nil {
				return fmt.Errorf("could not start web profiler: %w", err)
			}
		case "heapprofile":
			slog.Debug("Heap profiling enabled.")
		case "cpuprofile":
			if err := startCPUProfiler(flagVal); err != nil {
				return fmt.Errorf("could not start CPU profiling: %w", err)
			}
		case "traceprofile":
			if err := startTraceProfiling(flagVal); err != nil {
				return fmt.Errorf("could not start trace profiling: %w", err)
			}
		default:
			return fmt.Errorf("unknown argument for profiling: %s=%s", flagKey, flagVal)
		}
	}

	return nil
}

func StopProfiling(flags ProfileFlags) error {
	for flagKey, flagVal := range flags {
		switch flagKey {
		case "heapprofile":
			heapFile, err := os.Create(flagVal)
			if err != nil {
				return fmt.Errorf("cannot create heap profile file: %w", err)
			}

			var ms runtime.MemStats

			runtime.ReadMemStats(&ms)
			printMemStats(&ms)

			if err = pprof.WriteHeapProfile(heapFile); err != nil {
				return fmt.Errorf("cannot write to heap profile file: %w", err)
			}

			if err = heapFile.Close(); err != nil {
				return fmt.Errorf("cannot close heap profile: %w", err)
			}

			slog.Debug("Wrote heap profile.", "file", flagVal)
		case "cpuprofile":
			pprof.StopCPUProfile()
		case "traceprofile":
			trace.Stop()
		}
	}

	return nil
}

// printMemStats and formatMemory functions are taken from golang-ci source
//
//nolint:lll
func printMemStats(stats *runtime.MemStats) {
	slog.Debug("Memory stats",
		"alloc", formatMemory(stats.Alloc), "total_alloc", formatMemory(stats.TotalAlloc), "sys", formatMemory(stats.Sys),
		"heap_alloc", formatMemory(stats.HeapAlloc), "heap_sys", formatMemory(stats.HeapSys),
		"heap_idle", formatMemory(stats.HeapIdle), "heap_released", formatMemory(stats.HeapReleased), "heap_in_use", formatMemory(stats.HeapInuse),
		"stack_in_use", formatMemory(stats.StackInuse), "stack_sys", formatMemory(stats.StackSys),
		"mspan_sys", formatMemory(stats.MSpanSys), "mcache_sys", formatMemory(stats.MCacheSys), "buck_hash_sys", formatMemory(stats.BuckHashSys),
		"gc_sys", formatMemory(stats.GCSys), "other_sys", formatMemory(stats.OtherSys),
		"mallocs_n", stats.Mallocs, "frees_n", stats.Frees, "heap_objects", stats.HeapObjects, "gc_cpu_fraction", stats.GCCPUFraction)
}

//nolint:varnamelen
func formatMemory(memBytes uint64) string {
	const kb = 1024

	const mb = kb * 1024

	if memBytes < kb {
		return fmt.Sprintf("%db", memBytes)
	}

	if memBytes < mb {
		return fmt.Sprintf("%dkb", memBytes/kb)
	}

	return fmt.Sprintf("%dmb", memBytes/mb)
}

func startWebProfiler(enable string) error {
	webui, err := strconv.ParseBool(enable)
	if err != nil {
		return fmt.Errorf("could not interpret webui value: %w", err)
	}

	if webui {
		go func() {
			for i := 6060; i < 6070; i++ {
				slog.Debug("Starting profiler web interface.", "address", "http://localhost:"+strconv.Itoa(i))

				err := http.ListenAndServe("localhost:"+strconv.Itoa(i), nil) // #nosec G114
				if err != nil {
					slog.Error("Could not start profiler web interface. Trying different port.")
				}
			}
		}()
	}

	return nil
}

func startCPUProfiler(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cannot create CPU profile file: %w", err)
	}

	if err := pprof.StartCPUProfile(f); err != nil {
		return fmt.Errorf("could not start CPU profiling: %w", err)
	}

	slog.Debug("CPU profiling enabled.")

	return nil
}

func startTraceProfiling(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cannot create trace profile file: %w", err)
	}

	if err = trace.Start(f); err != nil {
		return fmt.Errorf("could not start trace profiling: %w", err)
	}

	slog.Debug("Trace profiling enabled.")

	return nil
}
