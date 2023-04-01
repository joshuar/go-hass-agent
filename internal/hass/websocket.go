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
	conn    *websocket.Conn
	ReadCh  chan *WebsocketResponse
	WriteCh chan interface{}
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

func (ws *HassWebsocket) Read(ctx context.Context) {
	for {
		response := &WebsocketResponse{
			Success: true,
		}
		err := wsjson.Read(ctx, ws.conn, &response)
		if err != nil {
			log.Debug().Caller().Msg(err.Error())
		}
		select {
		case <-ctx.Done():
			log.Debug().Caller().
				Msg("Stopping reading from websocket.")
			close(ws.ReadCh)
			return
		case ws.ReadCh <- response:
		}
	}
}

func (ws *HassWebsocket) Write(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Debug().Caller().
				Msg("Stopping writing to websocket.")
			ws.conn.Close(websocket.StatusNormalClosure, "")
			close(ws.WriteCh)
			return
		case data := <-ws.WriteCh:
			writeCtx, writeCancel := context.WithTimeout(ctx, time.Minute)
			defer writeCancel()
			err := wsjson.Write(writeCtx, ws.conn, data)
			if err != nil {
				log.Debug().Caller().Msg(err.Error())
			}
		}
	}
}

func (ws *HassWebsocket) Close() {
	ws.conn.Close(websocket.StatusNormalClosure, "requested websocket close")
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
	ws := &HassWebsocket{
		conn:    conn,
		ReadCh:  make(chan *WebsocketResponse),
		WriteCh: make(chan interface{}),
	}
	go ws.Read(ctx)
	go ws.Write(ctx)
	return ws
}
