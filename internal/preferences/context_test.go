// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:containedctx,exhaustruct,paralleltest
package preferences

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmbedInContext(t *testing.T) {
	type args struct {
		ctx context.Context
		p   *Preferences
	}

	tests := []struct {
		args args
		name string
	}{
		{
			name: "default",
			args: args{ctx: context.TODO(), p: defaultPreferences()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EmbedInContext(tt.args.ctx, tt.args.p)
			_, ok := got.Value(cfgKey).(Preferences)
			assert.True(t, ok)
		})
	}
}

func TestFetchFromContext(t *testing.T) {
	prefs := &Preferences{
		DeviceName: "testDevice",
	}
	fullCtx := EmbedInContext(context.TODO(), prefs)

	type args struct {
		ctx context.Context
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "fullCtx",
			args: args{ctx: fullCtx},
			want: "testDevice",
		},
		{
			name: "emptyCtx",
			args: args{ctx: context.TODO()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FetchFromContext(tt.args.ctx)
			assert.Equal(t, tt.want, got.DeviceName)
		})
	}
}
