// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:errname,paralleltest,wrapcheck,wsl,nlreturn // structs are dual-purpose response and error
package hass

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

//nolint:errname
type failedResponse struct {
	Details *preferences.Hass
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

func TestRegisterDevice(t *testing.T) {
	testDevice := preferences.DefaultPreferences(filepath.Join(t.TempDir(), "preferences.toml")).Device

	registrationSuccess := &registrationResponse{
		Details: &preferences.Hass{
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
		device       *preferences.Device
		registration *preferences.Registration
	}
	tests := []struct {
		args    args
		want    *preferences.Hass
		name    string
		wantErr bool
	}{
		{
			name: "successful registration",
			args: args{
				device: testDevice,
				registration: &preferences.Registration{
					Server: regSuccessServer.URL,
					Token:  "aToken",
				},
			},
			want: registrationSuccess.Details,
		},
		{
			name: "invalid input",
			args: args{
				registration: &preferences.Registration{},
				device:       testDevice,
			},
			wantErr: true,
		},
		{
			name: "failed registration",
			args: args{
				registration: &preferences.Registration{
					Server: regFailServer.URL,
					Token:  "aToken",
				},
				device: testDevice,
			},
			want:    registrationFail.Details,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RegisterDevice(context.TODO(), tt.args.device, tt.args.registration)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterDevice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RegisterDevice() = %v, want %v", got, tt.want)
			}
		})
	}
}
