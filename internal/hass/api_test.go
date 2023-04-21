// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

type FakeRequest struct {
	fakeState string
}

func (f *FakeRequest) RequestType() RequestType {
	return RequestTypeUpdateSensorStates
}

func (f *FakeRequest) RequestData() interface{} {
	return struct {
		State string `json:"State"`
	}{
		State: f.fakeState,
	}
}

func (f *FakeRequest) ResponseHandler(b bytes.Buffer) {
	spew.Dump(b.Bytes())
}

func TestMarshalJSONUnencrypted(t *testing.T) {
	input := &FakeRequest{
		fakeState: "foo",
	}
	output, _ := json.Marshal(&struct {
		Type RequestType `json:"type"`
		Data interface{} `json:"data"`
	}{
		Type: RequestTypeUpdateSensorStates,
		Data: struct {
			State string `json:"State"`
		}{
			State: input.fakeState,
		},
	})
	b, _ := MarshalJSON(input, "")
	if !bytes.Equal(b, output) {
		t.Errorf("Expected %v but got %v", string(output), string(b))
	}
}
