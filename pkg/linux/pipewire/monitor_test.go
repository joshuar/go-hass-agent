// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package pipewire

import (
	"context"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testNodeAdded = `{"id":42,"type":"PipeWire:Interface:Node","info":{"state":"running"}}`
	testNodeOther = `{"id":43,"type":"PipeWire:Interface:Node","info":{"state":"idle"}}`
	testRemoval   = `{"id":43,"info":null}`
)

// recordEvents adds a listener to the monitor that accepts every event and
// returns a function that reports the events received so far.
func recordEvents(t *testing.T, m *Monitor) func() []Event {
	t.Helper()

	var (
		mu     sync.Mutex
		events []Event
	)
	ch := m.AddListener(func(*Event) bool { return true })
	go func() {
		for event := range ch {
			mu.Lock()
			events = append(events, event)
			mu.Unlock()
		}
	}()

	return func() []Event {
		mu.Lock()
		defer mu.Unlock()

		return slices.Clone(events)
	}
}

func TestMonitorDecode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantIDs []int
		events  int
	}{
		{
			name:    "batches of events",
			input:   "[" + testNodeAdded + "," + testNodeOther + "]\n[" + testRemoval + "]\n",
			wantIDs: []int{42},
			events:  3,
		},
		{
			name:   "empty output",
			input:  "",
			events: 0,
		},
		{
			name:    "output ends mid batch",
			input:   "[" + testNodeAdded + ",",
			wantIDs: []int{42},
			events:  1,
		},
		{
			name:   "invalid output",
			input:  "not json",
			events: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			m := &Monitor{}
			events := recordEvents(t, m)

			done := make(chan map[int]struct{})
			go func() { done <- m.decode(ctx, strings.NewReader(tt.input)) }()

			var ids map[int]struct{}
			select {
			case ids = <-done:
			case <-time.After(time.Second):
				t.Fatal("decode did not return at end of output")
			}

			gotIDs := make([]int, 0, len(ids))
			for id := range ids {
				gotIDs = append(gotIDs, id)
			}
			require.ElementsMatch(t, tt.wantIDs, gotIDs)
			require.Eventually(t, func() bool { return len(events()) == tt.events }, time.Second, time.Millisecond)
		})
	}
}

func TestMonitorDispatchesRemovalsWhenPWDumpExits(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	m := &Monitor{}
	events := recordEvents(t, m)

	// Stand in for pw-dump: report two nodes and remove one, then exit.
	output := "[" + testNodeAdded + "," + testNodeOther + "]\n[" + testRemoval + "]\n"
	cmdCtx, cmdCancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(cmdCtx, "sh", "-c", `printf '%s' "$0"`, output)
	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)
	require.NoError(t, cmd.Start())

	m.monitor(ctx, &pwDump{cmd: cmd, stdout: stdout, cancel: cmdCancel})

	// The three reported events plus a removal for the node still present.
	require.Eventually(t, func() bool { return len(events()) == 4 }, time.Second, time.Millisecond)
	removals := slices.DeleteFunc(events(), func(e Event) bool { return !e.IsRemovalEvent() })
	ids := make([]int, 0, len(removals))
	for _, e := range removals {
		ids = append(ids, e.ID)
	}
	require.ElementsMatch(t, []int{43, 42}, ids)
}
