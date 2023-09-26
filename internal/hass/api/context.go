// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package api

import "context"

type APIConfig struct {
	APIURL string
	Secret string
}

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type key int

// userKey is the key for user.User values in Contexts. It is
// unexported; clients use user.NewContext and user.FromContext
// instead of using this key directly.
var userKey key

// NewContext returns a new Context that carries value u.
func NewContext(ctx context.Context, c *APIConfig) context.Context {
	return context.WithValue(ctx, userKey, c)
}

// FromContext returns the User value stored in ctx, if any.
func FromContext(ctx context.Context) (*APIConfig, bool) {
	u, ok := ctx.Value(userKey).(*APIConfig)
	return u, ok
}
