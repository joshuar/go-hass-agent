// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package cpu

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/joshuar/go-hass-agent/internal/linux"
)

func Test_loadAvgsWorker_Sensors(t *testing.T) {
	type fields struct {
		path     string
		loadAvgs []*linux.Sensor
	}
	type args struct {
		in0 context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{
			name:   "valid file",
			fields: fields{path: "testing/data/valid_loadavg", loadAvgs: newLoadAvgSensors()},
			want:   []string{"2.09", "2.14", "1.82"},
		},
		{
			name:    "invalid file",
			fields:  fields{path: "testing/data/nonexistent", loadAvgs: newLoadAvgSensors()},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &loadAvgsWorker{
				path:     tt.fields.path,
				loadAvgs: tt.fields.loadAvgs,
			}
			got, err := w.Sensors(tt.args.in0)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadAvgsWorker.Sensors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for idx, load := range tt.want {
				assert.Equal(t, load, got[idx].State().(string))
			}
		})
	}
}
