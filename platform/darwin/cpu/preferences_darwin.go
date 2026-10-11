package cpu

import "github.com/joshuar/go-hass-agent/agent/workers"

const prefPrefix = "sensors.cpu."

// UsagePrefs are the preferences for the CPU usage worker.
type UsagePrefs struct {
	workers.CommonWorkerPrefs `toml:",squash"`

	UpdateInterval string `toml:"update_interval"`
}

// LoadAvgPrefs are the preferences for the CPU load average worker.
type LoadAvgPrefs struct {
	workers.CommonWorkerPrefs `toml:",squash"`

	UpdateInterval string `toml:"update_interval"`
}
