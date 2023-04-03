package agent

type sensorState struct {
	deviceClass interface{}
	stateClass  string
	state       interface{}
	attributes  interface{}
	name        string
	entityID    string
	disabled    bool
	registered  bool
}
