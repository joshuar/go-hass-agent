package agent

func (agent *Agent) runSensorWorker() {

	go agent.runActiveAppSensor()

}

// func (a *Agent) updateDesktopInfo(d desktopInfo) error {
// 	spew.Dump(d.ActiveApp())
// 	return nil
// }

// func (agent *Agent) updateSensor(s *interface{}) error {

// }
