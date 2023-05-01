// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensors

import (
	"context"
	"reflect"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	badger "github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
)

func newMockSensorRegistry(t *testing.T) *badgerDBRegistry {
	fakeRegistry := new(badgerDBRegistry)
	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	assert.Nil(t, err)
	fakeRegistry.db = db
	return fakeRegistry
}

func TestOpenSensorRegistry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	assert.Nil(t, err)
	defer db.Close()
	wantedRegistry := &badgerDBRegistry{
		db: db,
	}
	badPath, _ := storage.ParseURI("file:///some/bad/path")
	type args struct {
		ctx          context.Context
		registryPath fyne.URI
	}
	tests := []struct {
		name    string
		args    args
		want    *badgerDBRegistry
		wantErr bool
	}{
		{
			name:    "successful open",
			args:    args{ctx: ctx, registryPath: nil},
			want:    wantedRegistry,
			wantErr: false,
		},
		{
			name:    "unsuccessful open",
			args:    args{ctx: ctx, registryPath: badPath},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := openSensorRegistry(tt.args.ctx, tt.args.registryPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("OpenSensorRegistry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				assert.IsType(t, wantedRegistry.db, got.db)
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("OpenSensorRegistry() = %v, want %v", got, tt.want)
			// }
		})
	}
}

func Test_sensorRegistry_CloseSensorRegistry(t *testing.T) {
	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	assert.Nil(t, err)
	defer db.Close()
	type fields struct {
		db *badger.DB
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "successful close",
			fields:  fields{db: db},
			wantErr: false,
		},
		{
			name:    "successful close on nonexistent db",
			fields:  fields{db: nil},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &badgerDBRegistry{
				db: tt.fields.db,
			}
			if err := reg.Close(); (err != nil) != tt.wantErr {
				t.Errorf("sensorRegistry.CloseSensorRegistry() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sensorRegistry_Get(t *testing.T) {
	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	assert.Nil(t, err)
	defer db.Close()
	reg := &badgerDBRegistry{
		db: db,
	}

	fakeMetadata := &sensorMetadata{
		Registered: true,
		Disabled:   false,
	}

	err = reg.Set("fakeSensor", fakeMetadata)
	assert.Nil(t, err)

	type fields struct {
		reg *badgerDBRegistry
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *sensorMetadata
		wantErr bool
	}{
		{
			name:   "existing",
			fields: fields{reg: reg},
			args:   args{id: "fakeSensor"},
			want:   fakeMetadata,
		},
		{
			name:    "nonexisting",
			fields:  fields{reg: reg},
			args:    args{id: "noSensor"},
			want:    &sensorMetadata{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := reg.Get(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("sensorRegistry.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorRegistry.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorRegistry_Set(t *testing.T) {
	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	assert.Nil(t, err)
	defer db.Close()

	fakeMetadata := &sensorMetadata{
		Registered: true,
		Disabled:   false,
	}

	type fields struct {
		db *badger.DB
	}
	type args struct {
		id     string
		values *sensorMetadata
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "add valid data",
			fields: fields{db: db},
			args:   args{id: "fakeSensor", values: fakeMetadata},
		},
		{
			name:   "add defaults",
			fields: fields{db: db},
			args:   args{id: "fakeSensor", values: &sensorMetadata{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &badgerDBRegistry{
				db: tt.fields.db,
			}
			if err := reg.Set(tt.args.id, tt.args.values); (err != nil) != tt.wantErr {
				t.Errorf("sensorRegistry.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
