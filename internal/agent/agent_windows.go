package agent

import (
	"context"
)

func Run() {
	ctx := context.Background()
	agent := NewAgent(ctx)
	agent.App.Run()
	agent.Exit()
}
