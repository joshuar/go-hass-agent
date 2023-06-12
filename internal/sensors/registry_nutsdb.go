// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensors

import (
	"context"
	"encoding/json"
	"os"

	"fyne.io/fyne/v2"
	"github.com/nutsdb/nutsdb"
)

const (
	registryBucket = "sensorRegistryV1"
)

type nutsdbRegistry struct {
	db *nutsdb.DB
}

func (r *nutsdbRegistry) Open(ctx context.Context, registryPath fyne.URI) error {
	var path string
	if registryPath != nil {
		path = registryPath.Path()
	} else {
		tmpDir, err := os.MkdirTemp("", "go-hass-agent-")
		if err != nil {
			return err
		}
		path = tmpDir
	}
	db, err := nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir(path),
	)
	if err != nil {
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

func (r *nutsdbRegistry) Get(id string) (*registryItem, error) {
	item := &registryItem{
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
		return item, err
	}
	return item, nil
}

func (r *nutsdbRegistry) Set(item registryItem) error {
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
