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

func Test_nutsdbRegistry_IsDisabled(t *testing.T) {
	mockDB := NewNutsDB(context.Background(), "")
	defer mockDB.Close()
	err := mockDB.SetDisabled("isDisabled", true)
	assert.Nil(t, err)
	err = mockDB.SetDisabled("notDisabled", false)
	assert.Nil(t, err)

	type fields struct {
		db *nutsdb.DB
	}
	type args struct {
		id string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "disabled sensor",
			fields: fields{
				db: mockDB.db,
			},
			args: args{
				id: "isDisabled",
			},
			want: true,
		},
		{
			name: "not disabled sensor",
			fields: fields{
				db: mockDB.db,
			},
			args: args{
				id: "notDisabled",
			},
			want: false,
		},
		{
			name: "not existing",
			fields: fields{
				db: mockDB.db,
			},
			args: args{
				id: "notExisting",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &nutsdbRegistry{
				db: tt.fields.db,
			}
			if got := r.IsDisabled(tt.args.id); got != tt.want {
				t.Errorf("nutsdbRegistry.IsDisabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_nutsdbRegistry_IsRegistered(t *testing.T) {
	mockDB := NewNutsDB(context.Background(), "")
	defer mockDB.Close()
	err := mockDB.SetRegistered("isRegistered", true)
	assert.Nil(t, err)
	err = mockDB.SetRegistered("notRegistered", false)
	assert.Nil(t, err)

	type fields struct {
		db *nutsdb.DB
	}
	type args struct {
		id string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "registered sensor",
			fields: fields{
				db: mockDB.db,
			},
			args: args{
				id: "isRegistered",
			},
			want: true,
		},
		{
			name: "not registered sensor",
			fields: fields{
				db: mockDB.db,
			},
			args: args{
				id: "notRegistered",
			},
			want: false,
		},
		{
			name: "not existing",
			fields: fields{
				db: mockDB.db,
			},
			args: args{
				id: "notExisting",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &nutsdbRegistry{
				db: tt.fields.db,
			}
			if got := r.IsRegistered(tt.args.id); got != tt.want {
				t.Errorf("nutsdbRegistry.IsRegistered() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_nutsdbRegistry_SetDisabled(t *testing.T) {
	mockDB := NewNutsDB(context.Background(), "")
	defer mockDB.Close()
	err := mockDB.SetDisabled("existingSensor", true)
	assert.Nil(t, err)

	type fields struct {
		db *nutsdb.DB
	}
	type args struct {
		id    string
		state bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "new sensor",
			fields: fields{
				db: mockDB.db,
			},
			args: args{
				id:    "newSensor",
				state: true,
			},
			wantErr: false,
		},
		{
			name: "existing sensor",
			fields: fields{
				db: mockDB.db,
			},
			args: args{
				id:    "existingSensor",
				state: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &nutsdbRegistry{
				db: tt.fields.db,
			}
			if err := r.SetDisabled(tt.args.id, tt.args.state); (err != nil) != tt.wantErr {
				t.Errorf("nutsdbRegistry.SetDisabled() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_nutsdbRegistry_SetRegistered(t *testing.T) {
	mockDB := NewNutsDB(context.Background(), "")
	defer mockDB.Close()
	err := mockDB.SetRegistered("existingSensor", true)
	assert.Nil(t, err)

	type fields struct {
		db *nutsdb.DB
	}
	type args struct {
		id    string
		state bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "new sensor",
			fields: fields{
				db: mockDB.db,
			},
			args: args{
				id:    "newSensor",
				state: true,
			},
			wantErr: false,
		},
		{
			name: "existing sensor",
			fields: fields{
				db: mockDB.db,
			},
			args: args{
				id:    "existingSensor",
				state: true,
			},
			wantErr: false,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &nutsdbRegistry{
				db: tt.fields.db,
			}
			if err := r.SetRegistered(tt.args.id, tt.args.state); (err != nil) != tt.wantErr {
				t.Errorf("nutsdbRegistry.SetRegistered() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_nutsdbRegistry_get(t *testing.T) {
	mockDB := NewNutsDB(context.Background(), "")
	defer mockDB.Close()
	err := mockDB.set("existingSensor", &SensorMetadata{
		Disabled:   true,
		Registered: true,
	})
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
		want    *SensorMetadata
		wantErr bool
	}{
		{
			name: "existing sensor",
			fields: fields{
				db: mockDB.db,
			},
			args: args{
				id: "existingSensor",
			},
			want: &SensorMetadata{
				Disabled:   true,
				Registered: true,
			},
			wantErr: false,
		},
		{
			name: "nonexisting sensor",
			fields: fields{
				db: mockDB.db,
			},
			args: args{
				id: "nonExistingSensor",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &nutsdbRegistry{
				db: tt.fields.db,
			}
			got, err := r.get(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("nutsdbRegistry.get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("nutsdbRegistry.get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_nutsdbRegistry_set(t *testing.T) {
	mockDB := NewNutsDB(context.Background(), "")
	defer mockDB.Close()
	err := mockDB.set("existingSensor", &SensorMetadata{
		Disabled:   true,
		Registered: false,
	})
	assert.Nil(t, err)

	type fields struct {
		db *nutsdb.DB
	}
	type args struct {
		id   string
		data *SensorMetadata
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "existing sensor",
			fields: fields{
				db: mockDB.db,
			},
			args: args{
				id: "existingSensor",
				data: &SensorMetadata{
					Disabled:   false,
					Registered: true,
				},
			},
			wantErr: false,
		},
		{
			name: "nonexisting sensor",
			fields: fields{
				db: mockDB.db,
			},
			args: args{
				id: "nonExistingSensor",
				data: &SensorMetadata{
					Disabled:   false,
					Registered: true,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &nutsdbRegistry{
				db: tt.fields.db,
			}
			if err := r.set(tt.args.id, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("nutsdbRegistry.set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewNutsDB(t *testing.T) {
	type args struct {
		ctx  context.Context
		path string
	}
	tests := []struct {
		name string
		args args
		want *nutsdbRegistry
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewNutsDB(tt.args.ctx, tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNutsDB() = %v, want %v", got, tt.want)
			}
		})
	}
}
