// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:exhaustruct,paralleltest,wsl
//revive:disable:unused-receiver
package sensor

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSensorTracker_add(t *testing.T) {
	type fields struct {
		sensor map[string]Details
	}
	type args struct {
		s Details
	}
	tests := []struct {
		args    args
		fields  fields
		name    string
		wantErr bool
	}{
		{
			name: "successful add",
			fields: fields{
				sensor: make(map[string]Details),
			},
			args:    args{s: &mockSensor},
			wantErr: false,
		},
		{
			name:    "unsuccessful add (not initialised properly)",
			args:    args{s: &mockSensor},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Tracker{
				sensor: tt.fields.sensor,
			}
			if err := tr.add(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("SensorTracker.add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSensorTracker_Get(t *testing.T) {
	mockMap := make(map[string]Details)
	mockMap["mock_sensor"] = &mockSensor

	type fields struct {
		sensor map[string]Details
	}

	type args struct {
		id string
	}

	tests := []struct {
		want    Details
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "successful get",
			fields:  fields{sensor: mockMap},
			args:    args{id: "mock_sensor"},
			wantErr: false,
			want:    &mockSensor,
		},
		{
			name:    "unsuccessful get",
			fields:  fields{sensor: mockMap},
			args:    args{id: "doesntExist"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Tracker{
				sensor: tt.fields.sensor,
			}

			got, err := tr.Get(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("SensorTracker.Get() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SensorTracker.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSensorTracker_SensorList(t *testing.T) {
	mockMap := make(map[string]Details)
	mockMap["mock_sensor"] = &mockSensor

	type fields struct {
		sensor map[string]Details
	}

	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name:   "with sensors",
			fields: fields{sensor: mockMap},
			want:   []string{"mock_sensor"},
		},
		{
			name: "without sensors",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Tracker{
				sensor: tt.fields.sensor,
			}
			if got := tr.SensorList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SensorTracker.SensorList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSensorTracker(t *testing.T) {
	tests := []struct {
		want    *Tracker
		name    string
		wantErr bool
	}{
		{
			name: "default test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTracker()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSensorTracker() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
		})
	}
}

//nolint:containedctx
func TestMergeSensorCh(t *testing.T) {
	testCtx, cancelFunc := context.WithCancel(context.TODO())
	go func() {
		time.Sleep(5 * time.Second)
		cancelFunc()
	}()

	sensor1 := mockSensor
	sensor1.IDFunc = func() string {
		return "sensor1"
	}
	ch1 := make(chan Details)
	go func() {
		ch1 <- &sensor1
		close(ch1)
	}()

	sensor2 := mockSensor
	sensor1.IDFunc = func() string {
		return "sensor2"
	}
	ch2 := make(chan Details)
	go func() {
		ch2 <- &sensor2
		close(ch2)
	}()

	type args struct {
		ctx      context.Context
		sensorCh []<-chan Details
	}

	tests := []struct {
		name string
		args args
		want []Details
	}{
		{
			name: "send two sensors in two different channels",
			args: args{ctx: testCtx, sensorCh: []<-chan Details{ch1, ch2}},
			want: []Details{&sensor1, &sensor2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeSensorCh(tt.args.ctx, tt.args.sensorCh...)
			var want []Details
			for s := range got {
				want = append(want, s)
			}
			if !reflect.DeepEqual(want, tt.want) {
				t.Errorf("MergeSensorCh() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSensorTracker_Reset(t *testing.T) {
	mockMap := make(map[string]Details)
	mockMap["mock_sensor"] = &mockSensor

	type fields struct {
		sensor map[string]Details
	}
	tests := []struct {
		fields fields
		name   string
	}{
		{
			name:   "default",
			fields: fields{sensor: mockMap},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Tracker{
				sensor: tt.fields.sensor,
			}
			tr.Reset()
			assert.Nil(t, tr.sensor)
		})
	}
}
