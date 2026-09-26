package temperature

import "github.com/joshuar/go-hass-agent/agent/workers"

const temperatureWorkerPreferencesID = "sensors.temperature"

type WorkerPrefs struct {
	workers.CommonWorkerPrefs `toml:",squash"`

	UpdateInterval string `toml:"update_interval"`
}
