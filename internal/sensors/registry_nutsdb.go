// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensors

import (
	"context"
	"encoding/json"
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

func (r *nutsdbRegistry) Get(id string) (*RegistryItem, error) {
	item := &RegistryItem{
		id: id,
		data: &sensorMetadata{
			Registered: false,
			Disabled:   false,
		},
	}
	if err := r.db.View(
		func(tx *nutsdb.Tx) error {
			key := []byte(id)
			if e, err := tx.Get(registryBucket, key); err != nil {
				return err
			} else {
				err := json.Unmarshal(e.Value, item.data)
				return err
			}
		}); err != nil {
		return nil, err
	}
	return item, nil
}

func (r *nutsdbRegistry) Set(item RegistryItem) error {
	v, err := json.Marshal(item.data)
	if err != nil {
		return err
	}
	if err := r.db.Update(
		func(tx *nutsdb.Tx) error {
			if err := tx.Put(registryBucket, []byte(item.id), v, 0); err != nil {
				return err
			}
			return nil
		}); err != nil {
		return err
	}
	return nil
}
