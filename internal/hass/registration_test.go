// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:exhaustruct,paralleltest,wrapcheck
//revive:disable:unnecessary-stmt
package hass

import (
	"context"
	"encoding/json"
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
}

func (r *failedResponse) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &r.Details)
}

// setup creates a context using a test http client and server which will
// return the given response when the ExecuteRequest function is called.
var setupTestCtx = func(t *testing.T, response Response) context.Context {
	t.Helper()

	ctx := context.TODO()
	// load client
	client := resty.New().
		SetTimeout(1 * time.Second).
		AddRetryCondition(
			func(rr *resty.Response, _ error) bool {
				return rr.StatusCode() == http.StatusTooManyRequests
			},
		)
	ctx = ContextSetClient(ctx, client)
	// load server
	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, _ *http.Request) {
		var resp []byte
		var err error
		switch rType := response.(type) {
		case *registrationResponse:
			resp, err = json.Marshal(rType.Details)
		}
		if err != nil {
			_, err := responseWriter.Write(json.RawMessage(`{"success":false}`))
			if err != nil {
				t.Fatal(err)
			}
		} else {
			_, err := responseWriter.Write(resp)
			if err != nil {
				t.Fatal(err)
			}
		}
	}))
	ctx = ContextSetURL(ctx, server.URL)
	// return loaded context
	return ctx
}

//nolint:containedctx
func TestRegisterWithHass(t *testing.T) {
	registrationSuccess := &registrationResponse{
		Details: &RegistrationDetails{
			CloudhookURL: "someURL",
			RemoteUIURL:  "someURL",
			Secret:       "",
			WebhookID:    "someID",
		},
	}
	successCtx := setupTestCtx(t, registrationSuccess)

	registrationFail := &failedResponse{}
	failCtx := setupTestCtx(t, registrationFail)

	type args struct {
		ctx    context.Context
		input  *RegistrationInput
		device DeviceInfo
	}

	tests := []struct {
		args    args
		want    *RegistrationDetails
		name    string
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
			name: "invalid input",
			args: args{
				ctx:    successCtx,
				input:  &RegistrationInput{},
				device: mockDevInfo,
			},
			wantErr: true,
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

func TestRegistrationInput_Validate(t *testing.T) {
	type fields struct {
		Server           string
		Token            string
		IgnoreOutputURLs bool
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:   "valid",
			fields: fields{Server: "http://localhost:8123", Token: "anyString"},
		},
		{
			name:    "invalid: host",
			fields:  fields{Server: "localhost:8123", Token: "anyString"},
			wantErr: true,
		},
		{
			name:    "invalid: missing token",
			fields:  fields{Server: "http://localhost:8123"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &RegistrationInput{
				Server:           tt.fields.Server,
				Token:            tt.fields.Token,
				IgnoreOutputURLs: tt.fields.IgnoreOutputURLs,
			}
			if err := i.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("RegistrationInput.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
