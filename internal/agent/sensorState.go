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

type sensorUpdate interface {
	ID() string
	Type() string
	Value() interface{}
	ExtraValues() interface{}
}
