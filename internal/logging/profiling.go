// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package logging

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"strconv"

	"github.com/rs/zerolog/log"
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
			log.Debug().Msg("Heap profiling enabled.")
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

			log.Debug().Msgf("Heap profile written to %s", flagVal)
		case "cpuprofile":
			pprof.StopCPUProfile()
		case "traceprofile":
			trace.Stop()
		}
	}

	return nil
}

// printMemStats and formatMemory functions are taken from golang-ci source

func printMemStats(stats *runtime.MemStats) {
	log.Debug().Msgf("Mem stats: alloc=%s total_alloc=%s sys=%s "+
		"heap_alloc=%s heap_sys=%s heap_idle=%s heap_released=%s heap_in_use=%s "+
		"stack_in_use=%s stack_sys=%s "+
		"mspan_sys=%s mcache_sys=%s buck_hash_sys=%s gc_sys=%s other_sys=%s "+
		"mallocs_n=%d frees_n=%d heap_objects_n=%d gc_cpu_fraction=%.2f",
		formatMemory(stats.Alloc), formatMemory(stats.TotalAlloc), formatMemory(stats.Sys),
		formatMemory(stats.HeapAlloc), formatMemory(stats.HeapSys),
		formatMemory(stats.HeapIdle), formatMemory(stats.HeapReleased), formatMemory(stats.HeapInuse),
		formatMemory(stats.StackInuse), formatMemory(stats.StackSys),
		formatMemory(stats.MSpanSys), formatMemory(stats.MCacheSys), formatMemory(stats.BuckHashSys),
		formatMemory(stats.GCSys), formatMemory(stats.OtherSys),
		stats.Mallocs, stats.Frees, stats.HeapObjects, stats.GCCPUFraction)
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
				log.Debug().
					Msgf("Starting profiler web interface on localhost:" + strconv.Itoa(i))

				err := http.ListenAndServe("localhost:"+strconv.Itoa(i), nil) // #nosec G114
				if err != nil {
					log.Debug().Err(err).
						Msg("Trouble starting profiler, trying again.")
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

	log.Debug().Msg("CPU profiling enabled.")

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

	log.Debug().Msg("Trace profiling enabled.")

	return nil
}
