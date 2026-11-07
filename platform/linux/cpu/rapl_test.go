// Copyright 2026 opticon454
// SPDX-License-Identifier: MIT

package cpu

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
)

// writeDomain creates a fake RAPL domain directory.
func writeDomain(t *testing.T, root, dir, name string, energy, maxRange uint64) string {
	t.Helper()

	path := filepath.Join(root, dir)
	require.NoError(t, os.MkdirAll(path, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(path, "name"), []byte(name+"\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(path, "max_energy_range_uj"), []byte(strconv.FormatUint(maxRange, 10)+"\n"), 0o600))
	setEnergy(t, path, energy)

	return path
}

func setEnergy(t *testing.T, dir string, energy uint64) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "energy_uj"), []byte(strconv.FormatUint(energy, 10)+"\n"), 0o600))
}

func Test_raplPower(t *testing.T) {
	tests := []struct {
		name                   string
		previous, current, max uint64
		elapsed                time.Duration
		want                   float64
		wantOK                 bool
	}{
		{name: "steady", previous: 1_000_000, current: 51_000_000, max: 262_143_328_850, elapsed: time.Second, want: 50, wantOK: true},
		{name: "over ten seconds", previous: 0, current: 300_000_000, max: 262_143_328_850, elapsed: 10 * time.Second, want: 30, wantOK: true},
		{name: "counter wrapped", previous: 99_000_000, current: 9_000_000, max: 100_000_000, elapsed: time.Second, want: 10, wantOK: true},
		{name: "wrapped but range unknown", previous: 99_000_000, current: 9_000_000, max: 0, elapsed: time.Second},
		{name: "no time passed", previous: 1, current: 2, max: 10, elapsed: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := raplPower(tt.previous, tt.current, tt.max, tt.elapsed)
			assert.Equal(t, tt.wantOK, ok)
			assert.InDelta(t, tt.want, got, 0.001)
		})
	}
}

func Test_findRAPLDomains(t *testing.T) {
	root := t.TempDir()
	assert.Empty(t, findRAPLDomains(root), "no domains in an empty directory")

	writeDomain(t, root, "intel-rapl:0", "package-0", 1, 100)
	writeDomain(t, root, "intel-rapl:0:1", "uncore", 1, 100)

	domains := findRAPLDomains(root)
	require.Len(t, domains, 2)
	assert.Equal(t, "package-0", domains[0].name)
	assert.Equal(t, "0", domains[0].id)
	assert.Equal(t, "uncore", domains[1].name)
	assert.Equal(t, "0_1", domains[1].id)
}

func newTestRAPLWorker(root string, clock *time.Time) *raplWorker {
	return &raplWorker{
		WorkerMetadata:          models.SetWorkerMetadata("cpu_power", "CPU power metrics"),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{OutCh: make(chan models.Entity, 16)},
		prefs:                   &RAPLPrefs{},
		domains:                 findRAPLDomains(root),
		now:                     func() time.Time { return *clock },
		previous:                make(map[string]raplReading),
	}
}

func Test_raplWorker_Execute(t *testing.T) {
	root := t.TempDir()
	pkg := writeDomain(t, root, "intel-rapl:0", "package-0", 1_000_000, 1_000_000_000)
	clock := time.Unix(1000, 0)
	worker := newTestRAPLWorker(root, &clock)

	// The first run only records the counter.
	require.NoError(t, worker.Execute(context.Background()))
	assert.Empty(t, worker.OutCh)

	// 35 W for 10 s.
	setEnergy(t, pkg, 1_000_000+350_000_000)
	clock = clock.Add(10 * time.Second)
	require.NoError(t, worker.Execute(context.Background()))
	require.Len(t, worker.OutCh, 1)

	sensor, err := (<-worker.OutCh).AsSensor()
	require.NoError(t, err)
	assert.Equal(t, "cpu_rapl_0_power", string(sensor.UniqueID))
	assert.Equal(t, "CPU package-0 Power", string(sensor.Name))
	assert.Equal(t, "power", sensor.DeviceClass)
	assert.Equal(t, "W", string(sensor.UnitOfMeasurement))
	assert.InDelta(t, 35.0, sensor.State, 0.001)
}

func Test_raplWorker_unreadable(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can read any file")
	}

	root := t.TempDir()
	pkg := writeDomain(t, root, "intel-rapl:0", "package-0", 1, 100)
	require.NoError(t, os.Chmod(filepath.Join(pkg, "energy_uj"), 0o000))

	clock := time.Unix(1000, 0)
	worker := newTestRAPLWorker(root, &clock)

	require.NoError(t, worker.Execute(context.Background()))
	assert.True(t, worker.unreadable, "permission denied stops further reads")
	assert.Empty(t, worker.OutCh)
}

func Test_raplWorker_IsDisabled(t *testing.T) {
	clock := time.Unix(0, 0)
	worker := newTestRAPLWorker(t.TempDir(), &clock)
	assert.True(t, worker.IsDisabled(), "disabled without RAPL domains")

	root := t.TempDir()
	writeDomain(t, root, "intel-rapl:0", "package-0", 1, 100)
	worker = newTestRAPLWorker(root, &clock)
	assert.False(t, worker.IsDisabled())

	worker.prefs.Disabled = true
	assert.True(t, worker.IsDisabled(), "disabled in the preferences")
}
