// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tracker

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

func (r *nutsdbRegistry) IsDisabled(id string) bool {
	metadata, err := r.get(id)
	if err != nil {
		log.Debug().Err(err).
			Msgf("Could not retrieve disabled state for %s from registry.", id)
		return false
	}
	return metadata.Disabled
}

func (r *nutsdbRegistry) IsRegistered(id string) bool {
	metadata, err := r.get(id)
	if err != nil {
		log.Debug().Err(err).
			Msgf("Could not retrieve registered state for %s from registry.", id)
		return false
	}
	return metadata.Registered
}

func (r *nutsdbRegistry) SetDisabled(id string, state bool) error {
	metadata, err := r.get(id)
	if err != nil && errors.Is(err, nutsdb.ErrBucketNotFound) {
		return err
	}
	if metadata == nil {
		metadata = new(SensorMetadata)
	}
	metadata.Disabled = state
	return r.set(id, metadata)
}

func (r *nutsdbRegistry) SetRegistered(id string, state bool) error {
	metadata, err := r.get(id)
	if err != nil && errors.Is(err, nutsdb.ErrBucketNotFound) {
		return err
	}
	if metadata == nil {
		metadata = new(SensorMetadata)
	}
	metadata.Registered = state
	return r.set(id, metadata)
}

func (r *nutsdbRegistry) get(id string) (*SensorMetadata, error) {
	data := new(SensorMetadata)
	if err := r.db.View(
		func(tx *nutsdb.Tx) error {
			key := []byte(id)
			if e, err := tx.Get(registryBucket, key); err != nil {
				return err
			} else {
				return json.Unmarshal(e.Value, data)
			}
		}); err != nil {
		return nil, err
	}
	return data, nil
}

func (r *nutsdbRegistry) set(id string, data *SensorMetadata) error {
	v, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if err := r.db.Update(
		func(tx *nutsdb.Tx) error {
			if err := tx.Put(registryBucket, []byte(id), v, 0); err != nil {
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
