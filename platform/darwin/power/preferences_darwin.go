package power

import "github.com/joshuar/go-hass-agent/agent/workers"

const controlsPrefPrefix = "controls.power."

const (
	powerControlPreferencesID   = controlsPrefPrefix + "power_controls"
	inhibitControlPreferencesID = controlsPrefPrefix + "inhibit_controls"
)

// Prefs are the preferences shared by the power control workers.
type Prefs struct {
	workers.CommonWorkerPrefs `toml:",squash"`
}
