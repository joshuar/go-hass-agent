// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensors

import (
	"context"
	"encoding/json"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	badger "github.com/dgraph-io/badger/v4"
	"github.com/rs/zerolog/log"
)

type sensorRegistry struct {
	uri fyne.URI
	db  *badger.DB
}

func OpenSensorRegistry(ctx context.Context, appPath fyne.URI) *sensorRegistry {
	uri, err := storage.Child(appPath, "sensorRegistry")
	if err != nil {
		log.Error().Err(err).
			Msg("Unable open sensor registry path. Will not be able to track sensor status.")
		return nil
	}

	// Open a badgerDB with largely the default options, but trying to optimise
	// for lower memory usage as per:
	// https://dgraph.io/docs/badger/get-started/#memory-usage
	db, err := badger.Open(badger.DefaultOptions(uri.Path()).
		// * If the number of sensors is large, this might need adjustment.
		WithMemTableSize(12 << 20))
	if err != nil {
		log.Debug().Err(err).Msg("Could not open sensor registry DB.")
		return nil
	}

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Debug().Caller().Msg("Running GC on registry DB.")
			again:
				err := db.RunValueLogGC(0.7)
				if err == nil {
					goto again
				}
			case <-ctx.Done():
				log.Debug().Caller().Msg("Closing registry.")
				db.Close()

			}
		}
	}()

	return &sensorRegistry{
		uri: uri,
		db:  db,
	}
}

func (reg *sensorRegistry) CloseSensorRegistry() error {
	return reg.db.Close()
}

func (reg *sensorRegistry) Get(id string) (*sensorMetadata, error) {
	state := &sensorMetadata{
		Registered: false,
		Disabled:   false,
	}
	err := reg.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(id))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			err := json.Unmarshal(val, state)
			return err
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return state, err
	}
	return state, nil
}

func (reg *sensorRegistry) Set(id string, values *sensorMetadata) error {
	err := reg.db.Update(func(txn *badger.Txn) error {
		v, err := json.Marshal(values)
		if err != nil {
			return err
		}
		err = reg.db.Update(func(txn *badger.Txn) error {
			err = txn.Set([]byte(id), v)
			return err
		})
		return err
	})
	return err
}
