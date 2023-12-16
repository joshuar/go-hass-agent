// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package dbushelpers

import (
	"context"
	"reflect"
	"testing"

	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/assert"
)

func TestNewBus(t *testing.T) {
	type args struct {
		ctx context.Context
		t   dbusType
	}
	tests := []struct {
		name string
		args args
		want dbusType
	}{
		{
			name: "session bus",
			args: args{ctx: context.TODO(), t: SessionBus},
			want: SessionBus,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBus(tt.args.ctx, tt.args.t); !reflect.DeepEqual(got.busType, tt.want) || got.conn == nil {
				t.Errorf("NewBus() = %v, want %v", got.conn, tt.want)
			}
		})
	}
}

func TestNewBusRequest(t *testing.T) {
	ctx := Setup(context.TODO())
	bus, ok := getBus(ctx, SessionBus)
	assert.True(t, ok)
	wantRequest := &busRequest{
		bus: bus,
	}

	type args struct {
		ctx     context.Context
		busType dbusType
	}
	tests := []struct {
		name string
		args args
		want *busRequest
	}{
		{
			name: "session request",
			args: args{ctx: ctx, busType: SessionBus},
			want: wantRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBusRequest(tt.args.ctx, tt.args.busType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBusRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewBusRequest2(t *testing.T) {
	type args struct {
		ctx     context.Context
		busType dbusType
	}
	tests := []struct {
		name string
		args args
		want *busRequest
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBusRequest2(tt.args.ctx, tt.args.busType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBusRequest2() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_busRequest_Path(t *testing.T) {
	type fields struct {
		bus          *Bus
		eventHandler func(*dbus.Signal)
		path         dbus.ObjectPath
		event        string
		dest         string
		match        []dbus.MatchOption
	}
	type args struct {
		p dbus.ObjectPath
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *busRequest
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &busRequest{
				bus:          tt.fields.bus,
				eventHandler: tt.fields.eventHandler,
				path:         tt.fields.path,
				event:        tt.fields.event,
				dest:         tt.fields.dest,
				match:        tt.fields.match,
			}
			if got := r.Path(tt.args.p); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("busRequest.Path() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_busRequest_Match(t *testing.T) {
	type fields struct {
		bus          *Bus
		eventHandler func(*dbus.Signal)
		path         dbus.ObjectPath
		event        string
		dest         string
		match        []dbus.MatchOption
	}
	type args struct {
		m []dbus.MatchOption
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *busRequest
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &busRequest{
				bus:          tt.fields.bus,
				eventHandler: tt.fields.eventHandler,
				path:         tt.fields.path,
				event:        tt.fields.event,
				dest:         tt.fields.dest,
				match:        tt.fields.match,
			}
			if got := r.Match(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("busRequest.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_busRequest_Event(t *testing.T) {
	type fields struct {
		bus          *Bus
		eventHandler func(*dbus.Signal)
		path         dbus.ObjectPath
		event        string
		dest         string
		match        []dbus.MatchOption
	}
	type args struct {
		e string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *busRequest
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &busRequest{
				bus:          tt.fields.bus,
				eventHandler: tt.fields.eventHandler,
				path:         tt.fields.path,
				event:        tt.fields.event,
				dest:         tt.fields.dest,
				match:        tt.fields.match,
			}
			if got := r.Event(tt.args.e); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("busRequest.Event() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_busRequest_Handler(t *testing.T) {
	type fields struct {
		bus          *Bus
		eventHandler func(*dbus.Signal)
		path         dbus.ObjectPath
		event        string
		dest         string
		match        []dbus.MatchOption
	}
	type args struct {
		h func(*dbus.Signal)
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *busRequest
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &busRequest{
				bus:          tt.fields.bus,
				eventHandler: tt.fields.eventHandler,
				path:         tt.fields.path,
				event:        tt.fields.event,
				dest:         tt.fields.dest,
				match:        tt.fields.match,
			}
			if got := r.Handler(tt.args.h); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("busRequest.Handler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_busRequest_Destination(t *testing.T) {
	type fields struct {
		bus          *Bus
		eventHandler func(*dbus.Signal)
		path         dbus.ObjectPath
		event        string
		dest         string
		match        []dbus.MatchOption
	}
	type args struct {
		d string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *busRequest
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &busRequest{
				bus:          tt.fields.bus,
				eventHandler: tt.fields.eventHandler,
				path:         tt.fields.path,
				event:        tt.fields.event,
				dest:         tt.fields.dest,
				match:        tt.fields.match,
			}
			if got := r.Destination(tt.args.d); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("busRequest.Destination() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_busRequest_GetProp(t *testing.T) {
	type fields struct {
		bus          *Bus
		eventHandler func(*dbus.Signal)
		path         dbus.ObjectPath
		event        string
		dest         string
		match        []dbus.MatchOption
	}
	type args struct {
		prop string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    dbus.Variant
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &busRequest{
				bus:          tt.fields.bus,
				eventHandler: tt.fields.eventHandler,
				path:         tt.fields.path,
				event:        tt.fields.event,
				dest:         tt.fields.dest,
				match:        tt.fields.match,
			}
			got, err := r.GetProp(tt.args.prop)
			if (err != nil) != tt.wantErr {
				t.Errorf("busRequest.GetProp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("busRequest.GetProp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_busRequest_SetProp(t *testing.T) {
	type fields struct {
		bus          *Bus
		eventHandler func(*dbus.Signal)
		path         dbus.ObjectPath
		event        string
		dest         string
		match        []dbus.MatchOption
	}
	type args struct {
		prop  string
		value dbus.Variant
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &busRequest{
				bus:          tt.fields.bus,
				eventHandler: tt.fields.eventHandler,
				path:         tt.fields.path,
				event:        tt.fields.event,
				dest:         tt.fields.dest,
				match:        tt.fields.match,
			}
			if err := r.SetProp(tt.args.prop, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("busRequest.SetProp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_busRequest_GetData(t *testing.T) {
	type fields struct {
		bus          *Bus
		eventHandler func(*dbus.Signal)
		path         dbus.ObjectPath
		event        string
		dest         string
		match        []dbus.MatchOption
	}
	type args struct {
		method string
		args   []any
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *dbusData
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &busRequest{
				bus:          tt.fields.bus,
				eventHandler: tt.fields.eventHandler,
				path:         tt.fields.path,
				event:        tt.fields.event,
				dest:         tt.fields.dest,
				match:        tt.fields.match,
			}
			if got := r.GetData(tt.args.method, tt.args.args...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("busRequest.GetData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_busRequest_Call(t *testing.T) {
	type fields struct {
		bus          *Bus
		eventHandler func(*dbus.Signal)
		path         dbus.ObjectPath
		event        string
		dest         string
		match        []dbus.MatchOption
	}
	type args struct {
		method string
		args   []any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &busRequest{
				bus:          tt.fields.bus,
				eventHandler: tt.fields.eventHandler,
				path:         tt.fields.path,
				event:        tt.fields.event,
				dest:         tt.fields.dest,
				match:        tt.fields.match,
			}
			if err := r.Call(tt.args.method, tt.args.args...); (err != nil) != tt.wantErr {
				t.Errorf("busRequest.Call() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_busRequest_AddWatch(t *testing.T) {
	type fields struct {
		bus          *Bus
		eventHandler func(*dbus.Signal)
		path         dbus.ObjectPath
		event        string
		dest         string
		match        []dbus.MatchOption
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &busRequest{
				bus:          tt.fields.bus,
				eventHandler: tt.fields.eventHandler,
				path:         tt.fields.path,
				event:        tt.fields.event,
				dest:         tt.fields.dest,
				match:        tt.fields.match,
			}
			if err := r.AddWatch(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("busRequest.AddWatch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_busRequest_RemoveWatch(t *testing.T) {
	type fields struct {
		bus          *Bus
		eventHandler func(*dbus.Signal)
		path         dbus.ObjectPath
		event        string
		dest         string
		match        []dbus.MatchOption
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &busRequest{
				bus:          tt.fields.bus,
				eventHandler: tt.fields.eventHandler,
				path:         tt.fields.path,
				event:        tt.fields.event,
				dest:         tt.fields.dest,
				match:        tt.fields.match,
			}
			if err := r.RemoveWatch(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("busRequest.RemoveWatch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_dbusData_AsVariantMap(t *testing.T) {
	aMap := make(map[string]any)
	aMap["foo"] = "foo"
	validMap := make(map[string]dbus.Variant)
	validMap["foo"] = dbus.MakeVariant("foo")
	type fields struct {
		data any
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]dbus.Variant
	}{
		{
			name:   "empty data",
			fields: fields{data: nil},
			want:   nil,
		},
		{
			name:   "valid data",
			fields: fields{data: aMap},
			want:   validMap,
		},
		{
			name:   "invalid data",
			fields: fields{data: string("aString")},
			want:   make(map[string]dbus.Variant),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &dbusData{
				data: tt.fields.data,
			}
			if got := d.AsVariantMap(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dbusData.AsVariantMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dbusData_AsStringMap(t *testing.T) {
	validMap := make(map[string]string)
	validMap["foo"] = "foo"
	type fields struct {
		data any
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]string
	}{
		{
			name:   "empty data",
			fields: fields{data: nil},
			want:   nil,
		},
		{
			name:   "valid data",
			fields: fields{data: validMap},
			want:   validMap,
		},
		{
			name:   "invalid data",
			fields: fields{data: string("aString")},
			want:   make(map[string]string),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &dbusData{
				data: tt.fields.data,
			}
			if got := d.AsStringMap(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dbusData.AsStringMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dbusData_AsObjectPathList(t *testing.T) {
	type fields struct {
		data any
	}
	tests := []struct {
		name   string
		fields fields
		want   []dbus.ObjectPath
	}{
		{
			name:   "empty data",
			fields: fields{data: nil},
			want:   nil,
		},
		{
			name:   "valid data",
			fields: fields{data: []dbus.ObjectPath{"/"}},
			want:   []dbus.ObjectPath{"/"},
		},
		{
			name:   "invalid data",
			fields: fields{data: string("aString")},
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &dbusData{
				data: tt.fields.data,
			}
			if got := d.AsObjectPathList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dbusData.AsObjectPathList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dbusData_AsStringList(t *testing.T) {
	type fields struct {
		data any
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name:   "empty data",
			fields: fields{data: nil},
			want:   nil,
		},
		{
			name:   "valid data",
			fields: fields{data: []string{"foo"}},
			want:   []string{"foo"},
		},
		{
			name:   "invalid data",
			fields: fields{data: string("aString")},
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &dbusData{
				data: tt.fields.data,
			}
			if got := d.AsStringList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dbusData.AsStringList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dbusData_AsObjectPath(t *testing.T) {
	type fields struct {
		data any
	}
	tests := []struct {
		name   string
		fields fields
		want   dbus.ObjectPath
	}{
		{
			name:   "empty data",
			fields: fields{data: nil},
			want:   dbus.ObjectPath(""),
		},
		{
			name:   "valid data",
			fields: fields{data: dbus.ObjectPath("/some/path")},
			want:   dbus.ObjectPath("/some/path"),
		},
		{
			name:   "invalid data",
			fields: fields{data: string("aString")},
			want:   dbus.ObjectPath(""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &dbusData{
				data: tt.fields.data,
			}
			if got := d.AsObjectPath(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dbusData.AsObjectPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dbusData_AsRawInterface(t *testing.T) {
	type fields struct {
		data any
	}
	tests := []struct {
		name   string
		fields fields
		want   any
	}{
		{
			name:   "empty data",
			fields: fields{data: nil},
			want:   nil,
		},
		{
			name:   "any data",
			fields: fields{data: string("")},
			want:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &dbusData{
				data: tt.fields.data,
			}
			if got := d.AsRawInterface(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dbusData.AsRawInterface() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSessionPath(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want dbus.ObjectPath
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSessionPath(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSessionPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVariantToValue(t *testing.T) {
	type args struct {
		variant dbus.Variant
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "string conversion",
			args: args{variant: dbus.MakeVariant("foo")},
			want: "foo",
		},
		// TODO: Test all possible variant values?
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VariantToValue[string](tt.args.variant); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VariantToValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
