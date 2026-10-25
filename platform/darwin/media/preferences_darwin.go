package media

import "github.com/joshuar/go-hass-agent/agent/workers"

const volumeControlPreferencesID = "controls.media.volume_control"

// VolumeWorkerPrefs are the preferences for the volume control worker.
type VolumeWorkerPrefs struct {
	workers.CommonWorkerPrefs `toml:",squash"`
}
