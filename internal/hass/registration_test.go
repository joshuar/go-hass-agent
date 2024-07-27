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
)

var mockDevInfo = &DeviceInfo{
	DeviceID:           "mockDeviceID",
	DeviceName:         "testDevice",
	Manufacturer:       "ACME",
	Model:              "Foobar",
	OsName:             "Fake OS",
	OsVersion:          "0.0.0",
	AppID:              "go-hass-agent-test",
	AppName:            "Go Hass Agent Test",
	AppVersion:         "v0.0.0",
	SupportsEncryption: false,
}

//nolint:errname
type failedResponse struct {
	Details *RegistrationDetails
	*APIError
}

func (r *failedResponse) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &r.Details)
}

func (r *failedResponse) UnmarshalError(b []byte) error {
	return json.Unmarshal(b, r.APIError)
}

func (r *failedResponse) Error() string {
	return r.APIError.Error()
}

// setup creates a context using a test http client and server which will
// return the given response when the ExecuteRequest function is called.
var setupTestServer = func(t *testing.T, response Response) *httptest.Server {
	t.Helper()
	// load server
	return httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, _ *http.Request) {
		var resp []byte
		var err error
		switch rType := response.(type) {
		case *registrationResponse:
			resp, err = json.Marshal(rType.Details)
		case *failedResponse:
			err = ErrUnknown
		}
		if err != nil {
			responseWriter.WriteHeader(http.StatusBadRequest)
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
}

//nolint:containedctx
//revive:disable:function-length
func TestRegisterWithHass(t *testing.T) {
	registrationSuccess := &registrationResponse{
		Details: &RegistrationDetails{
			CloudhookURL: "someURL",
			RemoteUIURL:  "someURL",
			Secret:       "",
			WebhookID:    "someID",
		},
	}

	regSuccessServer := setupTestServer(t, registrationSuccess)

	registrationFail := &failedResponse{}
	regFailServer := setupTestServer(t, registrationFail)

	type args struct {
		ctx    context.Context
		input  *RegistrationInput
		device *DeviceInfo
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
				ctx: context.TODO(),
				input: &RegistrationInput{
					Server: regSuccessServer.URL,
					Token:  "aToken",
				},
				device: mockDevInfo,
			},
			want: registrationSuccess.Details,
		},
		{
			name: "invalid input",
			args: args{
				ctx:    context.TODO(),
				input:  &RegistrationInput{},
				device: mockDevInfo,
			},
			wantErr: true,
		},
		{
			name: "failed registration",
			args: args{
				ctx: context.TODO(),
				input: &RegistrationInput{
					Server: regFailServer.URL,
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
