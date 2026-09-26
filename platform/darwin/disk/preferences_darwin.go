package disk

import "github.com/joshuar/go-hass-agent/agent/workers"

const (
	prefPrefix               = "sensors.disk."
	usageWorkerPreferencesID = prefPrefix + "usage"
	smartWorkerPreferencesID = prefPrefix + "smart"
)

type WorkerPrefs struct {
	workers.CommonWorkerPrefs `toml:",squash"`

	UpdateInterval string `toml:"update_interval"`
}
