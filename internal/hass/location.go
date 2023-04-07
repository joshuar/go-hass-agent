package hass

// LocationUpdate represents a location update from a platform/device. It
// provides a bridge between the platform/device specific location info and Home
// Assistant.
type LocationUpdate interface {
	Gps() []float64
	GpsAccuracy() int
	Battery() int
	Speed() int
	Altitude() int
	Course() int
	VerticalAccuracy() int
}

type locationUpdateInfo struct {
	Gps              []float64 `json:"gps"`
	GpsAccuracy      int       `json:"gps_accuracy,omitempty"`
	Battery          int       `json:"battery,omitempty"`
	Speed            int       `json:"speed,omitempty"`
	Altitude         int       `json:"altitude,omitempty"`
	Course           int       `json:"course,omitempty"`
	VerticalAccuracy int       `json:"vertical_accuracy,omitempty"`
}

func MarshalLocationUpdate(l LocationUpdate) *locationUpdateInfo {
	return &locationUpdateInfo{
		Gps:              l.Gps(),
		GpsAccuracy:      l.GpsAccuracy(),
		Battery:          l.Battery(),
		Speed:            l.Speed(),
		Altitude:         l.Altitude(),
		Course:           l.Course(),
		VerticalAccuracy: l.VerticalAccuracy(),
	}
}
