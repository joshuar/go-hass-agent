package hass

import (
	"context"
	"sync"
	"time"

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

func (conn *Conn) processRequests() {
	var wg sync.WaitGroup
	for request := range conn.requestsCh {
		wg.Add(1)
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		go func(request interface{}) {
			defer cancel()
			defer wg.Done()
			reqJson, err := MarshalJSON(request.(Request))
			if err != nil {
				log.Error().Msgf("Unable to format request: %v", err)
				conn.responsesCh <- nil
			} else {
				var res interface{}
				err = requests.
					URL(conn.requestURL).
					BodyBytes(reqJson).
					ToJSON(&res).
					Fetch(ctx)
				if err != nil {
					log.Error().Msgf("Unable to send request: %v", err)
					conn.responsesCh <- nil
				} else {
					conn.responsesCh <- res
				}
			}
		}(request)
	}
	wg.Wait()
}

func (conn *Conn) SendRequest(request Request) interface{} {
	conn.requestsCh <- request
	return <-conn.responsesCh
}
