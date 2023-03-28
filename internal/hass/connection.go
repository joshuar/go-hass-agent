package hass

import (
	"context"
	"sync"

	"github.com/carlmjohnson/requests"
	"github.com/rs/zerolog/log"
)

type Conn struct {
	requestsCh, responsesCh chan interface{}
	requestURL              string
}

func NewConnection(requestURL string) *Conn {
	newConn := &Conn{
		requestsCh:  make(chan interface{}),
		responsesCh: make(chan interface{}),
		requestURL:  requestURL,
	}
	go newConn.processRequests()
	return newConn
}

func (c *Conn) processRequests() {
	var wg sync.WaitGroup
	for request := range c.requestsCh {
		wg.Add(1)
		go func(request interface{}) {
			ctx := context.Background()
			defer wg.Done()
			reqJson, err := MarshalJSON(request.(Request))
			if err != nil {
				log.Error().Msgf("Unable to format request: %v", err)
				c.responsesCh <- nil
			} else {
				var res interface{}
				err = requests.
					URL(c.requestURL).
					BodyBytes(reqJson).
					ToJSON(&res).
					Fetch(ctx)
				if err != nil {
					log.Error().Msgf("Unable to send request: %v", err)
					c.responsesCh <- nil
				} else {
					c.responsesCh <- res
				}
			}
		}(request)
	}
	wg.Wait()
}

func (c *Conn) SendRequest(request Request) interface{} {
	c.requestsCh <- request
	return <-c.responsesCh
}
