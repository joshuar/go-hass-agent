package hass

import (
	"runtime"

	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/rs/zerolog/log"
)

type LocationInfo interface {
	Gps() []float64
	GpsAccuracy() int
	Battery() int
	Speed() int
	Altitude() int
	Course() int
	VerticalAccuracy() int
}

func RunLocationUpdater() {

	locationInfoCh := make(chan interface{})

	switch os := runtime.GOOS; os {
	case "linux":
		go linux.LinuxLocationUpdater(locationInfoCh)
	default:
		log.Error().Msg("Unsupported Operating System.")
	}

	for loc := range locationInfoCh {
		log.Debug().Caller().
			Msgf("Got location %v", loc.(LocationInfo).Gps())

	}
}
