package config

import "context"

type AppConfig struct {
	APIURL       string `json:"restapi_url"`
	WebSocketURL string `json:"instance_url"`
	Secret       string
	Token        string
	WebhookID    string
}

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type key int

// configKey is the key for agent.AppConfig values in Contexts. It is
// unexported; clients use user.NewContext and user.FromContext
// instead of using this key directly.
var configKey key

// NewContext returns a new Context that carries value c.
func NewContext(ctx context.Context, c *AppConfig) context.Context {
	return context.WithValue(ctx, configKey, c)
}

// FromContext returns the User value stored in ctx, if any.
func FromContext(ctx context.Context) (*AppConfig, bool) {
	c, ok := ctx.Value(configKey).(*AppConfig)
	return c, ok
}
