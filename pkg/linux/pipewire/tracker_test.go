// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package pipewire

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Events as emitted by pw-dump --monitor.
const (
	usbMicAdded = `{"id":63,"type":"PipeWire:Interface:Node","info":{"change-mask":["state","props"],` +
		`"state":"suspended","props":{"media.class":"Audio/Source","node.name":"usb_mic"}}}`
	usbMicRunning = `{"id":63,"type":"PipeWire:Interface:Node","info":{"change-mask":["state"],` +
		`"state":"running","props":{"media.class":"Audio/Source","node.name":"usb_mic"}}}`
	btMicAdded = `{"id":99,"type":"PipeWire:Interface:Node","info":{"change-mask":["state","props"],` +
		`"state":"running","props":{"media.class":"Audio/Source","node.name":"bt_mic"}}}`
	btMicSuspended = `{"id":99,"type":"PipeWire:Interface:Node","info":{"change-mask":["state"],` +
		`"state":"suspended","props":{"media.class":"Audio/Source","node.name":"bt_mic"}}}`
	// Older pw-dump versions omit props on a state-only change.
	usbMicRunningNoProps = `{"id":63,"type":"PipeWire:Interface:Node","info":{"change-mask":["state"],` +
		`"state":"running"}}`
	usbMicIdleNoProps = `{"id":63,"type":"PipeWire:Interface:Node","info":{"change-mask":["state"],` +
		`"state":"idle"}}`
	usbMicRemoved = `{"id":63,"info":null}`
	streamRunning = `{"id":84,"type":"PipeWire:Interface:Node","info":{"change-mask":["state","props"],` +
		`"state":"running","props":{"media.class":"Stream/Input/Audio","node.name":"chrome"}}}`
	portEvent = `{"id":67,"type":"PipeWire:Interface:Port","info":{"change-mask":["params"]}}`
)

func decodeEvent(t *testing.T, raw string) *Event {
	t.Helper()

	var event Event
	require.NoError(t, json.Unmarshal([]byte(raw), &event))

	return &event
}

func TestNodeTracker(t *testing.T) {
	type step struct {
		event      string
		wantPassed bool
		wantInUse  bool
	}

	tests := []struct {
		name  string
		steps []step
	}{
		{
			name: "another source suspending does not clear a running source",
			steps: []step{
				{event: usbMicAdded, wantPassed: true, wantInUse: false},
				{event: btMicAdded, wantPassed: true, wantInUse: true},
				{event: usbMicRunning, wantPassed: true, wantInUse: true},
				{event: btMicSuspended, wantPassed: true, wantInUse: true},
			},
		},
		{
			name: "state changes without props are applied to known nodes",
			steps: []step{
				{event: usbMicAdded, wantPassed: true, wantInUse: false},
				{event: usbMicRunningNoProps, wantPassed: true, wantInUse: true},
				{event: usbMicIdleNoProps, wantPassed: true, wantInUse: false},
			},
		},
		{
			name: "state changes without props are ignored for unknown nodes",
			steps: []step{
				{event: usbMicRunningNoProps, wantPassed: false, wantInUse: false},
			},
		},
		{
			name: "removing a running source clears it",
			steps: []step{
				{event: usbMicRunning, wantPassed: true, wantInUse: true},
				{event: usbMicRemoved, wantPassed: true, wantInUse: false},
				{event: usbMicRemoved, wantPassed: false, wantInUse: false},
			},
		},
		{
			name: "other media classes and object types are ignored",
			steps: []step{
				{event: streamRunning, wantPassed: false, wantInUse: false},
				{event: portEvent, wantPassed: false, wantInUse: false},
			},
		},
		{
			name: "repeated events without a state change are filtered out",
			steps: []step{
				{event: usbMicRunning, wantPassed: true, wantInUse: true},
				{event: usbMicRunning, wantPassed: false, wantInUse: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewNodeTracker(MediaAudioSource)
			for i, s := range tt.steps {
				assert.Equal(t, s.wantPassed, tracker.Filter(decodeEvent(t, s.event)), "step %d: filter result", i)
				assert.Equal(t, s.wantInUse, tracker.AnyIn(StateRunning), "step %d: in use", i)
			}
		})
	}
}
