// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package config

import (
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	"github.com/rs/zerolog/log"
)

type FyneConfig struct {
	prefs fyne.Preferences
}

func NewFyneConfig() *FyneConfig {
	return &FyneConfig{
		prefs: fyne.CurrentApp().Preferences(),
	}
}

// FyneConfig satisfies agent.AgentConfig

func (c *FyneConfig) Get(key string, value interface{}) error {
	switch v := value.(type) {
	case *string:
		*v = c.prefs.StringWithFallback(key, "NOTSET")
		if *v == "NOTSET" {
			return fmt.Errorf("key %s not set", key)
		}
	case *bool:
		*v = c.prefs.BoolWithFallback(key, false)
	default:
		return errors.New("unsupported config value")
	}
	return nil
}

func (c *FyneConfig) Set(key string, value interface{}) error {
	switch v := value.(type) {
	case string:
		c.prefs.SetString(key, v)
	case bool:
		c.prefs.SetBool(key, v)
	default:
		return errors.New("unsupported config value")
	}
	return nil
}

func (c *FyneConfig) Delete(key string) error {
	log.Debug().Msg("Not implemented.")
	return nil
}

func (c *FyneConfig) StoragePath(id string) (string, error) {
	agent := fyne.CurrentApp()
	rootPath := agent.Storage().RootURI()
	extraPath, err := storage.Child(rootPath, id)
	if err != nil {
		return "", err
	} else {
		return extraPath.Path(), nil
	}
}
