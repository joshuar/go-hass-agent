// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
)

var mockDevInfo = &DeviceInfoMock{
	DeviceIDFunc:           func() string { return "mockDeviceID" },
	DeviceNameFunc:         func() string { return "testDevice" },
	SupportsEncryptionFunc: func() bool { return false },
	AppDataFunc:            func() any { return nil },
	ManufacturerFunc:       func() string { return "ACME" },
	ModelFunc:              func() string { return "Foobar" },
	OsNameFunc:             func() string { return "Fake OS" },
	OsVersionFunc:          func() string { return "0.0" },
	AppIDFunc:              func() string { return "go-hass-agent-test" },
	AppNameFunc:            func() string { return "Go Hass Agent Test" },
	AppVersionFunc:         func() string { return "v0.0.0" },
}

type failedResponse struct {
	Details *RegistrationDetails
	err     error
}

func (r *failedResponse) StoreError(e error) {
	r.err = e
}

func (r *failedResponse) Error() string {
	return r.err.Error()
}

func (r *failedResponse) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &r.Details)
}

// setup creates a context using a test http client and server which will
// return the given response when the ExecuteRequest function is called.
var setupTestCtx = func(r Response) context.Context {
	ctx := context.TODO()
	// load client
	client := resty.New().
		SetTimeout(1 * time.Second).
		AddRetryCondition(
			func(rr *resty.Response, err error) bool {
				return rr.StatusCode() == http.StatusTooManyRequests
			},
		)
	ctx = ContextSetClient(ctx, client)
	// load server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var resp []byte
		var err error
		switch rType := r.(type) {
		case *registrationResponse:
			resp, err = json.Marshal(rType.Details)
		}
		if err != nil {
			w.Write(json.RawMessage(`{"success":false}`))
		} else {
			w.Write(resp)
		}
	}))
	ctx = ContextSetURL(ctx, server.URL)
	// return loaded context
	return ctx
}

func TestRegisterWithHass(t *testing.T) {
	registrationSuccess := &registrationResponse{
		Details: &RegistrationDetails{
			CloudhookURL: "someURL",
			RemoteUIURL:  "someURL",
			Secret:       "",
			WebhookID:    "someID",
		},
	}
	successCtx := setupTestCtx(registrationSuccess)

	registrationFail := &failedResponse{
		err: errors.New("response failed"),
	}
	failCtx := setupTestCtx(registrationFail)

	type args struct {
		ctx    context.Context
		input  *RegistrationInput
		device DeviceInfo
	}
	tests := []struct {
		name    string
		args    args
		want    *RegistrationDetails
		wantErr bool
	}{
		{
			name: "successful registration",
			args: args{
				ctx: successCtx,
				input: &RegistrationInput{
					Server: ContextGetURL(successCtx),
					Token:  "aToken",
				},
				device: mockDevInfo,
			},
			want: registrationSuccess.Details,
		},
		{
			name: "failed registration",
			args: args{
				ctx: failCtx,
				input: &RegistrationInput{
					Server: ContextGetURL(failCtx),
					Token:  "aToken",
				},
				device: mockDevInfo,
			},
			want:    registrationFail.Details,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RegisterWithHass(tt.args.ctx, tt.args.input, tt.args.device)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterWithHass() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RegisterWithHass() = %v, want %v", got, tt.want)
			}
		})
	}
}
