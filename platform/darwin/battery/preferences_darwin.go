package battery

import "github.com/joshuar/go-hass-agent/agent/workers"

const batteryWorkerPreferencesID = "sensors.battery"

type WorkerPrefs struct {
	workers.CommonWorkerPrefs `toml:",squash"`

	UpdateInterval string `toml:"update_interval"`
}
