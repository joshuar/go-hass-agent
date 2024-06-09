// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

type contextKey string

const (
	urlContextKey    contextKey = "url"
	clientContextKey contextKey = "client"
)

func ContextSetURL(ctx context.Context, url string) context.Context {
	newCtx := context.WithValue(ctx, urlContextKey, url)

	return newCtx
}

func ContextGetURL(ctx context.Context) string {
	url, ok := ctx.Value(urlContextKey).(string)
	if !ok {
		return ""
	}

	return url
}

func ContextSetClient(ctx context.Context, client *resty.Client) context.Context {
	newCtx := context.WithValue(ctx, clientContextKey, client)

	return newCtx
}

func ContextGetClient(ctx context.Context) *resty.Client {
	url, ok := ctx.Value(clientContextKey).(*resty.Client)
	if !ok {
		return nil
	}

	return url
}

func NewContext() (context.Context, context.CancelFunc) {
	prefs, err := preferences.Load()
	if err != nil {
		log.Warn().Err(err).Msg("Could not create context.")

		return nil, nil
	}

	baseCtx, cancelFunc := context.WithCancel(context.Background())
	hassCtx := ContextSetURL(baseCtx, prefs.RestAPIURL)
	hassCtx = ContextSetClient(hassCtx, NewDefaultHTTPClient())

	return hassCtx, cancelFunc
}
