// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cmd

import (
	"fmt"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"runtime/trace"

	"github.com/adrg/xdg"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/joshuar/go-hass-agent/cmd/text"
	"github.com/joshuar/go-hass-agent/internal/agent"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/registry"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

var (
	traceFlag    bool
	debugFlag    bool
	AppID        string
	profileFlag  bool
	cpuProfile   string
	heapProfile  string
	headlessFlag bool
	traceProfile string
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "go-hass-agent",
	Short: "A Home Assistant, native app integration for desktop/laptop devices.",
	Long:  text.RootCmdLongText,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logging.SetLoggingLevel(traceFlag, debugFlag, profileFlag)
		logging.SetLogFile("go-hass-agent.log")
		if cpuProfile != "" {
			f, err := os.Create(cpuProfile)
			if err != nil {
				log.Fatal().Err(err).Msg("Cannot create CPU profile.")
			}
			if err := pprof.StartCPUProfile(f); err != nil {
				log.Fatal().Err(err).Msg("Could not start CPU profiling.")
			}
		}
		if traceProfile != "" {
			f, err := os.Create(traceProfile)
			if err != nil {
				log.Fatal().Err(err).Msg("Cannot create trace profile.")
			}
			if err = trace.Start(f); err != nil {
				log.Fatal().Err(err).Msg("Could not start trace profiling.")
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		agent := agent.New(&agent.Options{
			Headless: headlessFlag,
			ID:       AppID,
		})
		var err error

		registry.SetPath(filepath.Join(xdg.ConfigHome, agent.AppID(), "sensorRegistry"))
		reg, err := registry.Load()
		if err != nil {
			log.Fatal().Err(err).Msg("Could not load sensor registry.")
		}

		preferences.SetPath(filepath.Join(xdg.ConfigHome, agent.AppID()))
		var trk *sensor.SensorTracker
		if trk, err = sensor.NewSensorTracker(); err != nil {
			log.Fatal().Err(err).Msg("Could not start sensor sensor.")
		}

		agent.Run(trk, reg)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if cpuProfile != "" {
			pprof.StopCPUProfile()
		}
		if heapProfile != "" {
			f, err := os.Create(heapProfile)
			if err != nil {
				log.Fatal().Err(err).Msg("Cannot create heap profile.")
			}

			var ms runtime.MemStats
			runtime.ReadMemStats(&ms)
			printMemStats(&ms)

			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Fatal().Err(err).Msg("Cannot write heap profile.")
			}
			_ = f.Close()
		}
		if traceProfile != "" {
			trace.Stop()
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Msg("Could not start.")
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&traceFlag, "trace", false,
		"trace output (default is false)")
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false,
		"debug output (default is false)")
	rootCmd.PersistentFlags().BoolVar(&profileFlag, "profile", false,
		"enable profiling (default is false)")
	rootCmd.PersistentFlags().StringVar(&cpuProfile, "cpu-profile", "",
		"write a CPU profile to specified file")
	rootCmd.PersistentFlags().StringVar(&heapProfile, "heap-profile", "",
		"write a heap profile to specified file")
	rootCmd.PersistentFlags().StringVar(&traceProfile, "trace-profile", "",
		"write a trace profile to specified file")
	rootCmd.PersistentFlags().StringVar(&AppID, "appid", "com.github.joshuar.go-hass-agent",
		"specify a custom app ID (for debugging)")
	rootCmd.PersistentFlags().BoolVar(&headlessFlag, "terminal", defaultHeadless(),
		"run in terminal (without a GUI)")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(registerCmd)
}

func defaultHeadless() bool {
	_, v := os.LookupEnv("DISPLAY")
	return !v
}

// printMemStats and formatMemory functions are taken from golang-ci source

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
