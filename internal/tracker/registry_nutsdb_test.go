// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tracker

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/nutsdb/nutsdb"
	"github.com/stretchr/testify/assert"
)

func Test_nutsdbRegistry_Open(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	type fields struct {
		db *nutsdb.DB
	}
	type args struct {
		ctx          context.Context
		registryPath string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "successful open",
			args:    args{ctx: ctx, registryPath: ""},
			wantErr: false,
		},
		{
			name:    "unsuccessful open",
			args:    args{ctx: ctx, registryPath: "/nonexistent"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &nutsdbRegistry{
				db: tt.fields.db,
			}
			if err := r.Open(tt.args.ctx, tt.args.registryPath); (err != nil) != tt.wantErr {
				t.Errorf("nutsdbRegistry.Open() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_nutsdbRegistry_Close(t *testing.T) {
	dname, err := os.MkdirTemp("", "sampledir")
	assert.Nil(t, err)
	defer os.RemoveAll(dname)
	db, err := nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir(dname),
	)
	assert.Nil(t, err)
	type fields struct {
		db *nutsdb.DB
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
			r := &nutsdbRegistry{
				db: tt.fields.db,
			}
			if err := r.Close(); (err != nil) != tt.wantErr {
				t.Errorf("nutsdbRegistry.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_nutsdbRegistry_Get(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := &nutsdbRegistry{}
	err := r.Open(ctx, "")
	assert.Nil(t, err)
	mockItem := NewRegistryItem("fakeSensor")
	mockItem.SetRegistered(true)
	err = r.Set(*mockItem)
	assert.Nil(t, err)

	type fields struct {
		db *nutsdb.DB
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *RegistryItem
		wantErr bool
	}{
		{
			name:   "existing",
			fields: fields{db: r.db},
			args:   args{id: "fakeSensor"},
			want:   mockItem,
		},
		{
			name:    "nonexisting",
			fields:  fields{db: r.db},
			args:    args{id: "noSensor"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &nutsdbRegistry{
				db: tt.fields.db,
			}
			got, err := r.Get(tt.args.id)
			spew.Dump(got)
			if (err != nil) != tt.wantErr {
				t.Errorf("nutsdbRegistry.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("nutsdbRegistry.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_nutsdbRegistry_Set(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := &nutsdbRegistry{}
	err := r.Open(ctx, "")
	assert.Nil(t, err)
	fakeMetadata := &SensorMetadata{
		Registered: true,
		Disabled:   false,
	}
	type fields struct {
		db *nutsdb.DB
	}
	type args struct {
		ID     string
		values *SensorMetadata
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "add valid data",
			fields: fields{db: r.db},
			args:   args{ID: "fakeSensor", values: fakeMetadata},
		},
		{
			name:   "add defaults",
			fields: fields{db: r.db},
			args:   args{ID: "fakeSensor", values: &SensorMetadata{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &nutsdbRegistry{
				db: tt.fields.db,
			}
			if err := r.Set(RegistryItem{ID: tt.args.ID, data: tt.args.values}); (err != nil) != tt.wantErr {
				t.Errorf("nutsdbRegistry.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
