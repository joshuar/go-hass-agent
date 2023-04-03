package agent

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/device"
)

func Run() {
	ctx := device.NewContextWithDeviceAPI(context.Background())
	agent := NewAgent(ctx)
	agent.App.Run()
	agent.Exit()
}
