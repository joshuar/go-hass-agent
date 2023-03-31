package hass

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
	HandleAPIResponse(interface{})
}

type sensorRegistrationInfo struct {
	StateAttributes   interface{} `json:"attributes,omitempty"`
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
	StateAttributes interface{} `json:"attributes,omitempty"`
	Icon            string      `json:"icon,omitempty"`
	State           interface{} `json:"state"`
	Type            string      `json:"type"`
	UniqueID        string      `json:"unique_id"`
}

func MarshallSensorData(s sensor) interface{} {
	if s.Registered() {
		return []sensorUpdateInfo{{
			StateAttributes: s.Attributes(),
			Icon:            s.Icon(),
			State:           s.State(),
			Type:            s.SensorType(),
			UniqueID:        s.UniqueID(),
		},
		}
	} else {
		return sensorRegistrationInfo{
			StateAttributes:   s.Attributes(),
			DeviceClass:       s.DeviceClass(),
			Icon:              s.Icon(),
			Name:              s.Name(),
			State:             s.State(),
			Type:              s.SensorType(),
			UniqueID:          s.UniqueID(),
			UnitOfMeasurement: s.UnitOfMeasurement(),
			StateClass:        s.StateClass(),
			EntityCategory:    s.EntityCategory(),
			Disabled:          s.Disabled(),
		}
	}
}
