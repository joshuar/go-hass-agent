// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package api

import (
	bytes "bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/settings"
	"github.com/stretchr/testify/assert"
)

func Test_marshalJSON(t *testing.T) {
	mockReq := &RequestMock{
		RequestDataFunc: func() json.RawMessage {
			return json.RawMessage(`{"someField": "someValue"}`)
		},
		RequestTypeFunc: func() RequestType {
			return RequestTypeUpdateSensorStates
		},
	}

	mockEncReq := &RequestMock{
		RequestDataFunc: func() json.RawMessage {
			return json.RawMessage(`{"someField": "someValue"}`)
		},
		RequestTypeFunc: func() RequestType {
			return RequestTypeEncrypted
		},
	}

	type args struct {
		request Request
		secret  string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "unencrypted request",
			args: args{request: mockReq},
			want: []byte(`{"type":"update_sensor_states","data":{"someField":"someValue"}}`),
		},
		{
			name:    "encrypted request without secret",
			args:    args{request: mockEncReq},
			want:    nil,
			wantErr: true,
		},
		{
			name: "encrypted request with secret",
			args: args{request: mockEncReq, secret: "fakeSecret"},
			want: []byte(`{"type":"encrypted","encrypted_data":{"someField":"someValue"},"encrypted":true}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := marshalJSON(tt.args.request, tt.args.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func mockServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := &UnencryptedRequest{}
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.Nil(t, err)
		switch req.Type {
		case "register_sensor":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success":true}`))
		case "encrypted":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success":true}`))
		}
	}))
}

type req struct {
	reqType RequestType
	data    json.RawMessage
}

func (r *req) RequestType() RequestType {
	return r.reqType
}

func (r *req) RequestData() json.RawMessage {
	return r.data
}

func (r *req) ResponseHandler(b bytes.Buffer, resp chan Response) {
	resp <- NewGenericResponse(nil, r.reqType)
}

type encReq struct {
}

func (r *encReq) RequestType() RequestType {
	return RequestTypeEncrypted
}

func (r *encReq) RequestData() json.RawMessage {
	return json.RawMessage(`{"someField": "someValue"}`)
}

func (r *encReq) ResponseHandler(b bytes.Buffer, resp chan Response) {
	resp <- NewGenericResponse(nil, RequestTypeEncrypted)

}

func TestExecuteRequest(t *testing.T) {
	server := mockServer(t)
	defer server.Close()

	goodSettings := settings.NewSettings()
	goodSettings.SetValue(settings.ApiURL, server.URL)
	goodSettings.SetValue(settings.Secret, "aSecret")
	goodCtx := settings.StoreInContext(context.Background(), goodSettings)

	badSettings := settings.NewSettings()
	badSettings.SetValue(settings.ApiURL, server.URL)
	badCtx := settings.StoreInContext(context.Background(), badSettings)

	type args struct {
		ctx        context.Context
		request    Request
		responseCh chan Response
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "good request",
			args: args{
				ctx:        goodCtx,
				request:    &req{reqType: RequestTypeRegisterSensor},
				responseCh: make(chan Response, 1),
			},
			wantErr: false,
		},
		{
			name: "bad encrypted request, missing secret",
			args: args{
				ctx:        badCtx,
				request:    &encReq{},
				responseCh: make(chan Response, 1),
			},
			wantErr: true,
		},
		{
			name: "good encrypted request",
			args: args{
				ctx:        goodCtx,
				request:    &encReq{},
				responseCh: make(chan Response, 1),
			},
			wantErr: false,
		},
		{
			name: "bad context, no config",
			args: args{
				ctx:        context.Background(),
				request:    &req{reqType: RequestTypeRegisterSensor},
				responseCh: make(chan Response, 1),
			},
			wantErr: true,
		},
		{
			name: "bad json",
			args: args{
				ctx: goodCtx,
				request: &req{
					reqType: RequestTypeRegisterSensor,
					data:    json.RawMessage(`sdgasghsdag`),
				},
				responseCh: make(chan Response, 1),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer close(tt.args.responseCh)
				resp := <-tt.args.responseCh
				if err := resp.Error(); (err != nil) != tt.wantErr {
					t.Errorf("api.TestExecuteRequest() error = %v, wantErr %v", err, tt.wantErr)
				}
			}()
			wg.Add(1)
			go func() {
				defer wg.Done()
				ExecuteRequest(tt.args.ctx, tt.args.request, tt.args.responseCh)
			}()
			wg.Wait()
		})
	}
}
