// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package settings

import (
	"context"
	"errors"
	"sync"
)

type Settings struct {
	mu     sync.RWMutex
	values map[string]string
}

func (s *Settings) GetValue(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v, ok := s.values[key]; !ok {
		return "", errors.New("not found")
	} else {
		return v, nil
	}
}

func (s *Settings) SetValue(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[key] = value
	return nil
}

func NewSettings() *Settings {
	return &Settings{
		mu:     sync.RWMutex{},
		values: make(map[string]string),
	}
}

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type key int

// contextKey is the key for Settings values in Contexts. It is
// unexported; clients use settings.StoreInContext and settings.FetchFromContext
// instead of using this key directly.
var contextKey key

// StoreInContext returns a new Context that stores the Config, c.
func StoreInContext(ctx context.Context, s *Settings) context.Context {
	return context.WithValue(ctx, contextKey, s)
}

// FetchFromContext returns the Config value stored in ctx, if any.
func FetchFromContext(ctx context.Context) (*Settings, error) {
	if c, ok := ctx.Value(contextKey).(*Settings); !ok {
		return nil, errors.New("no API in context")
	} else {
		return c, nil
	}
}
