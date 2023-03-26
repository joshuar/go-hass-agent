package agent

import "github.com/joshuar/go-hass-agent/internal/hass"

func (agent *Agent) runSensorWorker(conn *hass.Conn) {
	go agent.runActiveAppSensor(conn)
}
