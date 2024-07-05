// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:containedctx,exhaustruct,paralleltest,wsl,nlreturn
package preferences

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextSetPrefs(t *testing.T) {
	type args struct {
		ctx context.Context
		p   *Preferences
	}
	tests := []struct {
		args args
		want *Preferences
		name string
	}{
		{
			name: "set defaults",
			args: args{ctx: context.TODO(), p: DefaultPreferences()},
			want: DefaultPreferences(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContextSetPrefs(tt.args.ctx, tt.args.p)
			gotPrefs, ok := got.Value(prefsContextKey).(*Preferences)
			assert.True(t, ok)
			if !reflect.DeepEqual(gotPrefs, tt.want) {
				t.Errorf("ContextSetPrefs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContextGetPrefs(t *testing.T) {
	loadedCtx := ContextSetPrefs(context.TODO(), DefaultPreferences())

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		args    args
		want    *Preferences
		name    string
		wantErr bool
	}{
		{
			name: "loaded ctx",
			args: args{ctx: loadedCtx},
			want: DefaultPreferences(),
		},
		{
			name:    "empty ctx",
			args:    args{ctx: context.TODO()},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ContextGetPrefs(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ContextGetPrefs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContextGetPrefs() = %v, want %v", got, tt.want)
			}
		})
	}
}
