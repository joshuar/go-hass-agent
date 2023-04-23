// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensors

import (
	"context"
	"os"
	"testing"

	"fyne.io/fyne/v2/app"
	"github.com/stretchr/testify/assert"
)

func TestRegistry(t *testing.T) {
	testApp := app.NewWithID("org.joshuar.go-hass-agent-test")
	uri := testApp.Storage().RootURI()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Logf("Using %s for temporary registry DB for testing", uri.Path())

	t.Log("Test open registry")
	registry := OpenSensorRegistry(ctx, uri)
	assert.NotNil(t, registry)

	t.Log("Test close registry")
	err := registry.CloseSensorRegistry()
	assert.Nil(t, err)

	os.RemoveAll(registry.uri.Path())
}

func TestRegistryCancel(t *testing.T) {
	testApp := app.NewWithID("org.joshuar.go-hass-agent-test")
	uri := testApp.Storage().RootURI()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Logf("Using %s for temporary registry DB for testing", uri.Path())

	registry := OpenSensorRegistry(ctx, uri)
	assert.NotNil(t, registry)

	t.Log("Test handling cancel")
	cancel()

	os.RemoveAll(registry.uri.Path())
}

func TestRegistryAccess(t *testing.T) {
	testApp := app.NewWithID("org.joshuar.go-hass-agent-test")
	uri := testApp.Storage().RootURI()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Logf("Using %s for temporary registry DB for testing", uri.Path())

	registry := OpenSensorRegistry(ctx, uri)
	assert.NotNil(t, registry)

	data := &sensorMetadata{
		Registered: true,
		Disabled:   false,
	}

	t.Log("Test access with data")
	err := registry.Set("test", data)
	assert.Nil(t, err)

	got, err := registry.Get("test")
	assert.Nil(t, err)
	assert.Equal(t, data, got)

	t.Log("Test access without data")
	got, err = registry.Get("notExists")
	assert.NotNil(t, err)
	assert.NotEqual(t, got, data)

	os.RemoveAll(registry.uri.Path())
}
