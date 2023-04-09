// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

type location struct {
	data hass.LocationUpdate
}

// location implements hass.Request so its data can be sent to the HA API

func (l *location) RequestType() hass.RequestType {
	return hass.RequestTypeUpdateLocation
}

func (l *location) RequestData() interface{} {
	return hass.MarshalLocationUpdate(l.data)
}

func (l *location) ResponseHandler(rawResponse interface{}) {
	if rawResponse == nil {
		log.Debug().Caller().Msg("No response data.")
	} else {
		log.Debug().Caller().Msgf("Location updated to %v", l.data.Gps())
	}
}
