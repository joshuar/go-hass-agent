// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensors

import (
	"encoding/json"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	badger "github.com/dgraph-io/badger/v4"
	"github.com/rs/zerolog/log"
)

type SensorRegistry struct {
	uri fyne.URI
	db  *badger.DB
}

type sensorStates struct {
	Registered bool
	Disabled   bool
}

func OpenSensorRegistry(appPath fyne.URI) *SensorRegistry {
	uri, err := storage.Child(appPath, "sensorRegistry")
	if err != nil {
		log.Error().Err(err).
			Msg("Unable open sensor registry path. Will not be able to track sensor status.")
		return nil
	}

	db, err := badger.Open(badger.DefaultOptions(uri.Path()))
	if err != nil {
		log.Debug().Err(err).Msg("Could not open sensor registry DB.")
		return nil
	}
	return &SensorRegistry{
		uri: uri,
		db:  db,
	}
}

func (reg *SensorRegistry) CloseSensorRegistry() {
	reg.db.Close()
}

func (reg *SensorRegistry) NewState(sensor string) error {
	newState := &sensorStates{
		Registered: false,
		Disabled:   false,
	}
	v, err := json.Marshal(newState)
	if err != nil {
		return err
	}
	err = reg.db.Update(func(txn *badger.Txn) error {
		err = txn.Set([]byte(sensor), v)
		return err
	})
	return err
}

func (reg *SensorRegistry) GetState(sensor string) (*sensorStates, error) {
	state := &sensorStates{}
	err := reg.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(sensor))
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

func (reg *SensorRegistry) SetState(sensor, stateType string, value bool) error {
	currentState := &sensorStates{}
	err := reg.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(sensor))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			err := json.Unmarshal(val, currentState)
			return err
		})
		if err != nil {
			return err
		}
		switch stateType {
		case "registered":
			currentState.Registered = value
		case "disabled":
			currentState.Disabled = value
		}
		v, err := json.Marshal(currentState)
		if err != nil {
			return err
		}
		err = reg.db.Update(func(txn *badger.Txn) error {
			err = txn.Set([]byte(sensor), v)
			return err
		})
		return err
	})
	return err
}
