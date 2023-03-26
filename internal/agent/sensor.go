package agent

import "github.com/joshuar/go-hass-agent/internal/hass"

type sensor interface {
	Attributes() interface{}
	DeviceClass() string
	Icon() string
	Name() string
	State() interface{}
	SensorType() string
	UniqueID() string
	UnitOfMeasurement() string
	StateClass() string
	EntityCategory() string
	Disabled() bool
	Registered() bool
}

type sensorRequest struct {
	data      sensor
	encrypted bool
}

type sensorRegistrationInfo struct {
	Attributes        interface{} `json:"attributes,omitempty"`
	DeviceClass       string      `json:"device_class,omitempty"`
	Icon              string      `json:"icon,omitempty"`
	Name              string      `json:"name"`
	State             interface{} `json:"state"`
	Type              string      `json:"type"`
	UniqueID          string      `json:"unique_id"`
	UnitOfMeasurement string      `json:"unit_of_measurement,omitempty"`
	StateClass        string      `json:"state_class,omitempty"`
	EntityCategory    string      `json:"entity_category,omitempty"`
	Disabled          bool        `json:"disabled,omitempty"`
}

type sensorUpdateInfo struct {
	Attributes interface{} `json:"attributes,omitempty"`
	Icon       string      `json:"icon,omitempty"`
	State      interface{} `json:"state"`
	Type       string      `json:"type"`
	UniqueID   string      `json:"unique_id"`
}

func (s *sensorRequest) RequestType() hass.RequestType {
	if s.data.Registered() {
		return hass.RequestTypeUpdateSensorStates
	}
	return hass.RequestTypeRegisterSensor
}

func (s *sensorRequest) RequestData() interface{} {
	if s.data.Registered() {
		return []sensorUpdateInfo{{
			Attributes: s.data.Attributes(),
			Icon:       s.data.Icon(),
			State:      s.data.State(),
			Type:       s.data.SensorType(),
			UniqueID:   s.data.UniqueID(),
		},
		}
	} else {
		return sensorRegistrationInfo{
			Attributes:        s.data.Attributes(),
			DeviceClass:       s.data.DeviceClass(),
			Icon:              s.data.Icon(),
			Name:              s.data.Name(),
			State:             s.data.State(),
			Type:              s.data.SensorType(),
			UniqueID:          s.data.UniqueID(),
			UnitOfMeasurement: s.data.UnitOfMeasurement(),
			StateClass:        s.data.StateClass(),
			EntityCategory:    s.data.EntityCategory(),
			Disabled:          s.data.Disabled(),
		}
	}
}

func (s *sensorRequest) IsEncrypted() bool {
	return s.encrypted
}

