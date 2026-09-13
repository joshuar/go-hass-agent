package agent

import (
	"context"
	"log/slog"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/agent/workers/mqtt"
	"github.com/joshuar/go-hass-agent/models"
)

// darwinMQTTWorker represents the darwin-specific (macOS) MQTTWorker.
type darwinMQTTWorker struct {
	logger *slog.Logger
}

func (c *darwinMQTTWorker) Start(_ context.Context) (*mqtt.WorkerData, error) {
	return &mqtt.WorkerData{
		Configs:       []*models.MQTTConfig{},
		Subscriptions: []*models.MQTTSubscription{},
		Msgs:          make(chan models.MQTTMsg),
	}, nil
}

func (c *darwinMQTTWorker) IsDisabled() bool {
	return false
}

// CreateOSMQTTWorkers initializes the list of MQTT workers for sensors and
// returns those that are supported on this device.
func CreateOSMQTTWorkers(_ context.Context) (workers.MQTTWorker, error) {
	mqttController := &darwinMQTTWorker{}

	return mqttController, nil
}
