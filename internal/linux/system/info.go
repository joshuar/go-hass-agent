// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package system

import (
	"context"
	"sync"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

func InfoUpdater(_ context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details)

	// Get distribution name and version.
	distro, distroVersion := linux.GetDistroDetails()

	// Get kernel version.
	kernelVersion := linux.GetKernelVersion()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		sensorCh <- &linux.Sensor{
			SensorTypeValue: linux.SensorKernel,
			Value:           kernelVersion,
			IsDiagnostic:    true,
			IconString:      "mdi:chip",
			SensorSrc:       linux.DataSrcProcfs,
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		sensorCh <- &linux.Sensor{
			SensorTypeValue: linux.SensorDistribution,
			Value:           distro,
			IsDiagnostic:    true,
			IconString:      "mdi:linux",
			SensorSrc:       linux.DataSrcProcfs,
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		sensorCh <- &linux.Sensor{
			SensorTypeValue: linux.SensorVersion,
			Value:           distroVersion,
			IsDiagnostic:    true,
			IconString:      "mdi:numeric",
			SensorSrc:       linux.DataSrcProcfs,
		}
	}()
	go func() {
		defer close(sensorCh)
		wg.Wait()
	}()
	return sensorCh
}
