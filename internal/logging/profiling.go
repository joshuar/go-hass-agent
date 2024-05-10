// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package logging

import (
	"errors"
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

func StartProfiling(flags ProfileFlags) error {
	for k, v := range flags {
		switch k {
		case "webui":
			webui, err := strconv.ParseBool(v)
			if err != nil {
				return errors.Join(errors.New("could not interpret webui value"), err)
			}
			if webui {
				go func() {
					for i := 6060; i < 6070; i++ {
						log.Debug().
							Msgf("Starting profiler web interface on localhost:" + fmt.Sprint(i))
						err := http.ListenAndServe("localhost:"+fmt.Sprint(i), nil)
						if err != nil {
							log.Debug().Err(err).
								Msg("Trouble starting profiler, trying again.")
						}
					}
				}()
			}
		case "heapprofile":
			log.Debug().Msg("Heap profiling enabled.")
		case "cpuprofile":
			f, err := os.Create(v)
			if err != nil {
				log.Fatal().Err(err).Msg("Cannot create CPU profile.")
			}
			if err := pprof.StartCPUProfile(f); err != nil {
				log.Fatal().Err(err).Msg("Could not start CPU profiling.")
			}
			log.Debug().Msg("CPU profiling enabled.")
		case "traceprofile":
			f, err := os.Create(v)
			if err != nil {
				log.Fatal().Err(err).Msg("Cannot create trace profile.")
			}
			if err = trace.Start(f); err != nil {
				log.Fatal().Err(err).Msg("Could not start trace profiling.")
			}
			log.Debug().Msg("Trace profiling enabled.")
		default:
			return fmt.Errorf("unknown argument for profiling: %s=%s", k, v)
		}
	}
	return nil
}

func StopProfiling(flags ProfileFlags) error {
	for k, v := range flags {
		switch k {
		case "heapprofile":
			f, err := os.Create(v)
			if err != nil {
				return errors.Join(errors.New("cannot create heap profile file"), err)
			}
			var ms runtime.MemStats
			runtime.ReadMemStats(&ms)
			printMemStats(&ms)
			if err := pprof.WriteHeapProfile(f); err != nil {
				return errors.Join(errors.New("cannot write to heap profile file"), err)
			}
			_ = f.Close()
			log.Debug().Msgf("Heap profile written to %s", v)
		case "cpuprofile":
			pprof.StopCPUProfile()
		case "traceprofile":
			trace.Stop()
		}
	}
	return nil
}

// printMemStats and formatMemory functions are taken from golang-ci source

func printMemStats(ms *runtime.MemStats) {
	log.Debug().Msgf("Mem stats: alloc=%s total_alloc=%s sys=%s "+
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
