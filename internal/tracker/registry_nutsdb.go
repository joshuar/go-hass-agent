// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tracker

import (
	"context"
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
	item := NewRegistryItem(id)
	if err := r.db.View(
		func(tx *nutsdb.Tx) error {
			key := []byte(id)
			if e, err := tx.Get(registryBucket, key); err != nil {
				return err
			} else {
				return item.UnmarshalJSON(e.Value)
			}
		}); err != nil {
		return nil, err
	}
	return item, nil
}

func (r *nutsdbRegistry) Set(item RegistryItem) error {
	v, err := item.MarshalJSON()
	if err != nil {
		return err
	}
	if err := r.db.Update(
		func(tx *nutsdb.Tx) error {
			if err := tx.Put(registryBucket, []byte(item.ID), v, 0); err != nil {
				return err
			}
			return nil
		}); err != nil {
		return err
	}
	return nil
}

func NewNutsDB(ctx context.Context, path string) *nutsdbRegistry {
	r := &nutsdbRegistry{}
	err := r.Open(ctx, path)
	if err != nil {
		log.Debug().Err(err).Caller().
			Msg("Unable to open registry")
		return nil
	}
	return r
}
