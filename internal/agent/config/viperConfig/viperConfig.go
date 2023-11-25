// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package viperconfig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var (
	configFileType = "toml"
	configFileName = "go-hass-agent"
	configFile     = configFileName + "." + configFileType
)

type ViperConfig struct {
	store *viper.Viper
	path  string
}

// Get sets the passed in value parameter to the value of the passed in key. The
// passed in value needs to be one of the supported values, either a string or
// bool reference. If the key is not set or the type of the passed in value is
// unsupported, the returned error will be non-nil.
func (c *ViperConfig) Get(key string, value interface{}) error {
	switch v := value.(type) {
	case *string:
		*v = c.store.GetString(key)
		if *v == "" {
			return fmt.Errorf("key %s not set", key)
		}
	case *bool:
		*v = c.store.GetBool(key)
	default:
		return errors.New("unsupported config value")
	}
	return nil
}

// Set will assign the passed in value to the passed in key in the config. If
// there is a problem setting the value or the key does not exist, it will
// return a non-nil error.
func (c *ViperConfig) Set(key string, value interface{}) error {
	c.store.Set(key, value)
	if err := c.store.WriteConfigAs(filepath.Join(c.path, configFile)); err != nil {
		log.Error().Err(err).Msg("Problem writing config file.")
	}
	if !c.store.IsSet(key) {
		return errors.New("value not set")
	}
	return nil
}

// Delete is currently unimplemented for ViperConfig.
func (c *ViperConfig) Delete(key string) error {
	return nil
}

// StoragePath returns a full path on the filesystem whose trailing path will be
// the given id.
func (c *ViperConfig) StoragePath(id string) (string, error) {
	return filepath.Join(c.path, id), nil
}

// New sets up a new ViperConfig object at the given path. If there is already a
// config at the path, it is opened and returned. Otherwise a new config is
// initialised.
func New(path string) (*ViperConfig, error) {
	c := &ViperConfig{
		store: viper.New(),
		path:  path,
	}
	if err := createDir(c.path); err != nil {
		return nil, errors.New("could not create config directory")
	}
	c.store.SetConfigType(configFileType)
	c.store.SetConfigName(configFileName)
	c.store.AddConfigPath(c.path)
	if err := c.store.ReadInConfig(); err != nil && !errors.Is(err, err.(viper.ConfigFileNotFoundError)) {
		return nil, err
	}
	return c, nil
}

func createDir(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		log.Debug().Str("directory", path).Msg("No config directory, creating new one.")
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}
