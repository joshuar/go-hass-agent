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
	badger "github.com/dgraph-io/badger/v4"
	"github.com/rs/zerolog/log"
)

type sensorRegistry struct {
	db *badger.DB
}

func openSensorRegistry(ctx context.Context, registryPath fyne.URI) (*sensorRegistry, error) {
	var db *badger.DB
	var err error
	if registryPath != nil {
		// Open a badgerDB with largely the default options, but trying to
		// optimise for lower memory usage as per:
		// https://dgraph.io/docs/badger/get-started/#memory-usage
		db, err = badger.Open(badger.DefaultOptions(registryPath.Path()).
			// * If the number of sensors is large, this might need adjustment.
			WithMemTableSize(12 << 20))
		if err != nil {
			log.Debug().Err(err).Msg("Could not open sensor registry DB.")
			return nil, err
		}
	} else {
		// As a fallback when no registryPath was provided, open an in-memory
		// database.  This will allow the agent to continue working, but sensor
		// registered and disabled states will not be tracked.
		db, err = badger.Open(badger.DefaultOptions("").WithInMemory(true))
		if err != nil {
			log.Debug().Err(err).Msg("Could not open sensor registry DB.")
			return nil, err
		}
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

	return &sensorRegistry{db: db}, nil
}

func (reg *sensorRegistry) closeSensorRegistry() error {
	if reg.db != nil {
		return reg.db.Close()
	} else {
		return nil
	}
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
