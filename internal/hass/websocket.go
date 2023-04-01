package hass

import (
	"context"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/rs/zerolog/log"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type HassWebsocket struct {
	conn *websocket.Conn
	ctx  context.Context
}

type WebsocketResponse struct {
	Type    string `json:"type"`
	Success bool   `json:"success,omitempty"`
	Error   struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
	ID           int         `json:"id,omitempty"`
	Result       interface{} `json:"result,omitempty"`
	HAVersion    string      `json:"ha_version,omitempty"`
	Notification struct {
		Message   string      `json:"message"`
		Title     string      `json:"title,omitempty"`
		Target    []string    `json:"target,omitempty"`
		Data      interface{} `json:"data,omitempty"`
		ConfirmID string      `json:"confirm_id,omitempty"`
	} `json:"event,omitempty"`
}

func (ws *HassWebsocket) Read(responseCh chan WebsocketResponse) {
	for {
		ctx, cancel := context.WithCancel(ws.ctx)
		select {
		case <-ws.ctx.Done():
		case <-responseCh:
			cancel()
			return
		default:
			response := &WebsocketResponse{
				Success: true,
			}
			err := wsjson.Read(ctx, ws.conn, &response)
			responseCh <- *response
			if err != nil {
				log.Warn().Msg(err.Error())
			}
		}
	}
}

func (ws *HassWebsocket) Write(data interface{}) error {
	ctx, cancel := context.WithTimeout(ws.ctx, time.Minute)
	defer cancel()
	err := wsjson.Write(ctx, ws.conn, data)
	return err
}

func NewWebsocket(ctx context.Context, url string) *HassWebsocket {
	log.Debug().Caller().Msgf("Using %s for websocket connection.", url)
	ctxConnect, cancelConnect := context.WithTimeout(ctx, time.Minute)
	defer cancelConnect()
	var conn *websocket.Conn
	var err error
	retryFunc := func() error {
		conn, _, err = websocket.Dial(ctxConnect, url, nil)
		if err != nil {
			return err
		}
		return nil
	}
	err = backoff.Retry(retryFunc, backoff.WithContext(backoff.NewExponentialBackOff(), ctxConnect))
	if err != nil {
		log.Debug().Caller().Msg(err.Error())
		cancelConnect()
		return nil
	}
	return &HassWebsocket{
		conn: conn,
		ctx:  ctx,
	}
}
