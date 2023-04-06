package agent

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/device"
)

func Run() {
	ctx, cancelfunc := context.WithCancel(context.Background())
	deviceCtx := device.Init(ctx)
	RunAgent(deviceCtx)
	cancelfunc()
}
