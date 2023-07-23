// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package registry

import (
	"encoding/json"
	"os"

	"github.com/nutsdb/nutsdb"
	"github.com/rs/zerolog/log"
)

const (
	registryBucket = "sensorRegistryV1"
)

type metadata struct {
	Registered bool `json:"Registered"`
	Disabled   bool `json:"Disabled"`
}

func MigrateNuts2Json(path string) error {
	var ndb *nutsdb.DB
	ndb, err := nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir(path),
	)
	if err != nil {
		return err
	}
	if ndb == nil {
		log.Debug().Msg("No nutsdb found. Skipping migration.")
		return nil
	}

	nutsdbEntries, err := getAllNutsEntries(ndb)
	if err != nil {
		return err
	}

	for _, entry := range nutsdbEntries {
		var err error
		id := string(entry.Key)
		m := &metadata{}
		err = json.Unmarshal(entry.Value, &m)
		if err != nil {
			log.Warn().Err(err).
				Msgf("Problem unmarshaling metadata for %s.", id)
		}
		log.Debug().Msgf("Found %s with metadata %v", id, m)
		err = writeJsonFile(path, id, *m)
		if err != nil {
			log.Warn().Err(err).
				Msgf("Problem writing json metadata for %s.", id)
		} else {
			log.Debug().Msgf("Successfully migrated metadata for %s.", id)
		}
	}

	return nil
}

func getAllNutsEntries(db *nutsdb.DB) (nutsdb.Entries, error) {
	var entries nutsdb.Entries
	err := db.View(
		func(tx *nutsdb.Tx) error {
			e, err := tx.GetAll(registryBucket)
			if err != nil {
				return err
			}
			entries = e
			return nil
		})
	return entries, err
}

func writeJsonFile(path, id string, m metadata) error {
	filePath := path + "/" + id + ".json"
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, b, 0644)
}
