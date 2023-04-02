package agent

type sensorState struct {
	deviceClass interface{}
	state       interface{}
	attributes  interface{}
	name        string
	entityID    string
	disabled    bool
	registered  bool
}
