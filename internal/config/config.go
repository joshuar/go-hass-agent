// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type AppConfig struct {
	APIURL       string `validate:"required"`
	WebSocketURL string `validate:"required"`
	Secret       string
	Token        string `validate:"required"`
	WebhookID    string `validate:"required"`
	NotifyCh     chan fyne.Notification
}

type validationError struct {
	Namespace       string `json:"namespace"` // can differ when a custom TagNameFunc is registered or
	Field           string `json:"field"`     // by passing alt name to ReportError like below
	StructNamespace string `json:"structNamespace"`
	StructField     string `json:"structField"`
	Tag             string `json:"tag"`
	ActualTag       string `json:"actualTag"`
	Kind            string `json:"kind"`
	Type            string `json:"type"`
	Value           string `json:"value"`
	Param           string `json:"param"`
	Message         string `json:"message"`
}

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type key int

// configKey is the key for agent.AppConfig values in Contexts. It is
// unexported; clients use config.NewContext and config.FromContext
// instead of using this key directly.
var configKey key

// StoreConfigInContext returns a new Context that carries value c.
func StoreConfigInContext(ctx context.Context, c *AppConfig) context.Context {
	return context.WithValue(ctx, configKey, c)
}

// FetchConfigFromContext returns the value stored in ctx, if any.
func FetchConfigFromContext(ctx context.Context) (*AppConfig, error) {
	if c, ok := ctx.Value(configKey).(*AppConfig); !ok {
		return nil, errors.New("no API in context")
	} else {
		return c, nil
	}
}

func (config *AppConfig) Validate() error {
	validate := validator.New()
	err := validate.Struct(config)
	if err != nil {

		// this check is only needed when your code could produce
		// an invalid value for validation such as interface with nil
		// value most including myself do not usually have code like this.
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return err
		}

		for _, err := range err.(validator.ValidationErrors) {
			e := validationError{
				Namespace:       err.Namespace(),
				Field:           err.Field(),
				StructNamespace: err.StructNamespace(),
				StructField:     err.StructField(),
				Tag:             err.Tag(),
				ActualTag:       err.ActualTag(),
				Kind:            fmt.Sprintf("%v", err.Kind()),
				Type:            fmt.Sprintf("%v", err.Type()),
				Value:           fmt.Sprintf("%v", err.Value()),
				Param:           err.Param(),
				Message:         err.Error(),
			}

			_, err := json.MarshalIndent(e, "", "  ")
			if err != nil {
				panic(err)
			}
		}

		// from here you can create your own error messages in whatever language you wish
		log.Debug().Caller().Msg("Config seems invalid.")
		return errors.New("invalid config")
	}
	return nil
}
