// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"errors"
	"os"
	"time"

	"github.com/joshuar/go-hass-agent/internal/tracker/registry"
	"github.com/rs/zerolog/log"
	"golang.org/x/mod/semver"
)

func Upgrade(c *agentConfig) error {
	configVersion := c.prefs.String("Version")
	if configVersion == "" {
		return errors.New("config version is not a valid value")
	}

	switch {
	// * Upgrade host to include scheme for versions < v.1.4.0
	case semver.Compare(configVersion, "v1.4.0") < 0:
		log.Debug().Msg("Performing config upgrades for < v1.4.0")
		hostString := c.prefs.String("Host")
		if hostString == "" {
			return errors.New("upgrade < v.1.4.0: invalid host value")
		}
		switch c.prefs.Bool("UseTLS") {
		case true:
			hostString = "https://" + hostString
		case false:
			hostString = "http://" + hostString
		}
		c.prefs.SetString("Host", hostString)
		fallthrough
	// * Add ApiURL and WebSocketURL config options for versions < v1.4.3
	case semver.Compare(configVersion, "v1.4.3") < 0:
		log.Debug().Msg("Performing config upgrades for < v1.4.3")
		c.generateAPIURL()
		c.generateWebsocketURL()
	case semver.Compare(Version, "v3.0.0") < 0:
		log.Debug().Msg("Performing config upgrades for < v3.0.0.")
		var err error
		path, err := c.NewStorage("sensorRegistry")
		if err != nil {
			return errors.New("could not get sensor registry path from config")
		}
		if _, err := os.Stat(path + "/0.dat"); errors.Is(err, os.ErrNotExist) {
			return nil
		}
		err = registry.MigrateNuts2Json(path)
		if err != nil {
			return errors.New("failed to migrate sensor registry")
		}
		if err = os.Remove(path + "/0.dat"); err != nil {
			return errors.New("could not remove old sensor registry")
		}
	}

	c.prefs.SetString("Version", Version)

	// ! https://github.com/fyne-io/fyne/issues/3170
	time.Sleep(110 * time.Millisecond)

	return nil
}
