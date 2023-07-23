// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package registry

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/nutsdb/nutsdb"
	"github.com/rs/zerolog/log"
)

const (
	registryBucket = "sensorRegistryV1"
)

type SensorMetadata struct {
	Registered bool `json:"Registered"`
	Disabled   bool `json:"Disabled"`
}

type nutsdbRegistry struct {
	db *nutsdb.DB
}

func (r *nutsdbRegistry) Open(ctx context.Context, path string) error {
	if path == "" {
		tmpDir, err := os.MkdirTemp("", "go-hass-agent-")
		if err != nil {
			return err
		}
		path = tmpDir
	}
	if db, err := nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir(path),
	); err != nil {
		log.Debug().Msg("bad path")
		return err
	} else {
		r.db = db
	}
	return nil
}

func (r *nutsdbRegistry) Close() error {
	if r.db != nil {
		return r.db.Close()
	} else {
		return nil
	}
}

func (r *nutsdbRegistry) IsDisabled(id string) chan bool {
	valueCh := make(chan bool, 1)
	go r.get(id, "disabled", valueCh)
	return valueCh
}

func (r *nutsdbRegistry) IsRegistered(id string) chan bool {
	valueCh := make(chan bool, 1)
	go r.get(id, "registered", valueCh)
	return valueCh
}

func (r *nutsdbRegistry) SetDisabled(id string, value bool) error {
	return r.set(id, "disabled", value)
}

func (r *nutsdbRegistry) SetRegistered(id string, value bool) error {
	return r.set(id, "registered", value)
}

func (r *nutsdbRegistry) get(id, valueType string, valueCh chan bool) {
	defer close(valueCh)
	if err := r.db.View(
		func(tx *nutsdb.Tx) error {
			key := []byte(id)
			entry, err := tx.Get(registryBucket, key)
			if err != nil && errors.Is(err, nutsdb.ErrBucketNotFound) {
				return err
			}
			data := &SensorMetadata{}
			if entry == nil {
				return nil
			}
			if err := json.Unmarshal(entry.Value, &data); err != nil {
				return err
			}
			switch valueType {
			case "disabled":
				valueCh <- data.Disabled
			case "registered":
				valueCh <- data.Registered
			}
			return nil
		}); err != nil {
		log.Debug().Err(err).
			Msgf("Unable to retrieve %s value for %s.", valueType, id)
		return
	}
}

func (r *nutsdbRegistry) set(id, valueType string, value bool) error {
	if err := r.db.Update(
		func(tx *nutsdb.Tx) error {
			key := []byte(id)
			var entry *nutsdb.Entry
			var err error
			entry, err = tx.Get(registryBucket, key)
			if err != nil && errors.Is(err, nutsdb.ErrBucketNotFound) {
				return err
			}
			data := &SensorMetadata{}
			if entry != nil {
				if err = json.Unmarshal(entry.Value, &data); err != nil {
					return err
				}
			}
			switch valueType {
			case "disabled":
				data.Disabled = value
			case "registered":
				data.Registered = value
			}
			newData, err := json.Marshal(data)
			if err != nil {
				return err
			}
			if err := tx.Put(registryBucket, []byte(id), newData, nutsdb.Persistent); err != nil {
				return err
			}
			return nil
		}); err != nil {
		return err
	}
	return nil
}

func NewNutsDB(ctx context.Context, path string) (*nutsdbRegistry, error) {
	r := &nutsdbRegistry{}
	err := r.Open(ctx, path)
	if err != nil {
		return nil, err
	}
	return r, nil
}
