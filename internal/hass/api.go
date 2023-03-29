package hass

import (
	"context"
	"time"

	"github.com/carlmjohnson/requests"
	"github.com/cenkalti/backoff"
	"github.com/rs/zerolog/log"
)

func APIRequest(ctx context.Context, url string, request interface{}, response func(r interface{})) {

	requestCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	reqJson, err := MarshalJSON(request.(Request))
	if err != nil {
		log.Error().Msgf("Unable to format request: %v", err)
		response(nil)
	} else {
		var res interface{}
		requestFunc := func() error {
			return requests.
				URL(url).
				BodyBytes(reqJson).
				ToJSON(&res).
				Fetch(requestCtx)
		}
		err := backoff.Retry(requestFunc, backoff.NewExponentialBackOff())
		if err != nil {
			log.Error().Msgf("Unable to send request: %v", err)
			response(nil)
		} else {
			response(res)
		}
	}
}
