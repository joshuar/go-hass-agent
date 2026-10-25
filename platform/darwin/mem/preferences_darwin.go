package mem

import "github.com/joshuar/go-hass-agent/agent/workers"

const memUsagePreferencesID = "sensors.memory.usage"

// WorkerPrefs are the preferences for the memory usage worker.
type WorkerPrefs struct {
	workers.CommonWorkerPrefs `toml:",squash"`

	UpdateInterval string `toml:"update_interval"`
}
