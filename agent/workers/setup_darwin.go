package workers

import (
	"context"

	"github.com/joshuar/go-hass-agent/platform/darwin"
)

// SetupCtx sets up a context for darwin (macOS) systems.
func SetupCtx(ctx context.Context) context.Context {
	return darwin.NewContext(ctx)
}
