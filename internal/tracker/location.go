// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tracker

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/api"
)

func updateLocation(ctx context.Context, l *hass.LocationData) {
	response := <-api.ExecuteRequest(ctx, l)
	switch r := response.(type) {
	case []byte:
		log.Debug().Msg("Location Updated.")
	case error:
		log.Warn().Err(r).Msg("Failed to update location.")
	default:
		log.Warn().Msgf("Unknown response type %T", r)
	}
}
