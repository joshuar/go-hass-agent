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

func (c *ViperConfig) Delete(key string) error {
	return nil
}

func (c *ViperConfig) StoragePath(id string) (string, error) {
	return filepath.Join(c.path, id), nil
}

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
