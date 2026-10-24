// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package hass

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/hass/api"
	"github.com/joshuar/go-hass-agent/models"
)

// fakeRegistry is a minimal in-memory sensorRegistry for tests.
type fakeRegistry struct {
	registered map[string]bool
	disabled   map[string]bool
}

func (f *fakeRegistry) SetDisabled(id string, state bool) error {
	if f.disabled == nil {
		f.disabled = make(map[string]bool)
	}
	f.disabled[id] = state
	return nil
}

func (f *fakeRegistry) SetRegistered(id string, state bool) error {
	if f.registered == nil {
		f.registered = make(map[string]bool)
	}
	f.registered[id] = state
	return nil
}

func (f *fakeRegistry) IsDisabled(id string) bool {
	return f.disabled[id]
}

func (f *fakeRegistry) IsRegistered(id string) bool {
	return f.registered[id]
}

// fakeTracker is a minimal in-memory sensorTracker for tests.
type fakeTracker struct{}

func (f *fakeTracker) SensorList() []models.UniqueID {
	return nil
}

func (f *fakeTracker) Get(id models.UniqueID) (*models.Sensor, error) {
	return nil, nil
}

func (f *fakeTracker) Add(details *models.Sensor) error {
	return nil
}

// newTestClient returns a Client whose sensor updates fail fast (the API URL
// is unreachable) and whose pending retry interval is short.
func newTestClient() *Client {
	return &Client{
		sensorRegistry: &fakeRegistry{registered: map[string]bool{"power_state": true}},
		sensorTracker:  &fakeTracker{},
		config: &Config{
			APIURL: "http://127.0.0.1:1",
			remote: &api.ConfigResponse{},
		},
	}
}

// TestEntityHandler_FailingSendDoesNotBlockNextEntity verifies that a failing
// sensor update does not block the handler from processing the next entity.
// It uses an unbuffered channel (like the real MergeCh) and a timeout on the
// send of the second entity to detect a block.
func TestEntityHandler_FailingSendDoesNotBlockNextEntity(t *testing.T) {
	client := newTestClient()

	entityCh := make(chan models.Entity) // unbuffered, like MergeCh
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go client.EntityHandler(ctx, entityCh)

	entityA := models.NewSensor(ctx,
		models.WithName("Power State"),
		models.WithID("power_state"),
		models.WithState("Powered On"),
	)
	entityB := models.NewSensor(ctx,
		models.WithName("Link State"),
		models.WithID("link_state"),
		models.WithState("Connected"),
	)

	// Feed the first entity. It fails (unreachable API URL) and is stored
	// pending. If the handler blocked on the failed send, the send of the
	// second entity would block; the timeout detects that.
	entityCh <- entityA
	select {
	case entityCh <- entityB:
		// The handler accepted the second entity, so it did not block.
	case <-time.After(2 * time.Second):
		t.Fatal("handler blocked on a failed send and did not process the next entity")
	}

	// Both failed updates should be pending.
	require.Eventually(t, func() bool {
		client.pendingMu.Lock()
		defer client.pendingMu.Unlock()
		return len(client.pending) == 2
	}, 2*time.Second, 10*time.Millisecond)
}

// TestEntityHandler_RetriesFailedUpdateAfterServerComesBack verifies that a
// failed Power State update is resent once the server is reachable again.
func TestEntityHandler_RetriesFailedUpdateAfterServerComesBack(t *testing.T) {
	var mu sync.Mutex
	fail := true
	received := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if fail {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		received++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := newTestClient()
	client.config.APIURL = srv.URL
	client.retryInterval = 50 * time.Millisecond

	entityCh := make(chan models.Entity)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go client.EntityHandler(ctx, entityCh)

	entity := models.NewSensor(ctx,
		models.WithName("Power State"),
		models.WithID("power_state"),
		models.WithState("Powered On"),
	)

	// The first send fails (server returns 500) and is stored pending.
	entityCh <- entity

	// Bring the server back up.
	mu.Lock()
	fail = false
	mu.Unlock()

	// The retry ticker should resend the pending update and succeed.
	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return received >= 1
	}, 3*time.Second, 50*time.Millisecond)
}

// TestEntityHandler_NewerStateReplacesPendingOlderOne verifies that a newer
// update for the same sensor ID replaces a pending older one, so the latest
// state wins.
func TestEntityHandler_NewerStateReplacesPendingOlderOne(t *testing.T) {
	client := newTestClient()

	// Use a buffered channel so both entities are queued and processed in
	// order. An unbuffered channel does not guarantee the two distinct
	// entities are read in order, which would make the assertion flaky.
	entityCh := make(chan models.Entity, 2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go client.EntityHandler(ctx, entityCh)

	older := models.NewSensor(ctx,
		models.WithName("Power State"),
		models.WithID("power_state"),
		models.WithState("Powered On"),
	)
	newer := models.NewSensor(ctx,
		models.WithName("Power State"),
		models.WithID("power_state"),
		models.WithState("Powered Off"),
	)

	// Both fail (unreachable API URL). The pending map should hold only the
	// newer state for the sensor.
	entityCh <- older
	entityCh <- newer

	require.Eventually(t, func() bool {
		client.pendingMu.Lock()
		defer client.pendingMu.Unlock()
		return len(client.pending) == 1
	}, 2*time.Second, 10*time.Millisecond)

	client.pendingMu.Lock()
	defer client.pendingMu.Unlock()
	pendingEntity, ok := client.pending["power_state"]
	require.True(t, ok, "expected a pending entry for power_state")
	pendingSensor, err := pendingEntity.AsSensor()
	require.NoError(t, err)
	assert.Equal(t, "Powered Off", pendingSensor.State)
}

// TestEntityHandler_PendingOldStateNotSentAfterNewerSuccess verifies that a
// pending old state is not resent once a newer update for the same sensor has
// succeeded. Because every send happens on the single handler goroutine, a
// newer successful update clears the pending entry before any retry can send
// the older one.
func TestEntityHandler_PendingOldStateNotSentAfterNewerSuccess(t *testing.T) {
	var mu sync.Mutex
	requestCount := 0
	successfulRequests := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		_, _ = io.Copy(io.Discard, r.Body)
		requestCount++
		if requestCount == 1 {
			// First request fails so the old update is stored pending.
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		successfulRequests++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"power_state":{"success":true}}`))
	}))
	defer srv.Close()

	client := newTestClient()
	client.config.APIURL = srv.URL
	client.retryInterval = 100 * time.Millisecond

	entityCh := make(chan models.Entity, 2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go client.EntityHandler(ctx, entityCh)

	old := models.NewSensor(ctx,
		models.WithName("Power State"),
		models.WithID("power_state"),
		models.WithState("Powered On"),
	)
	newer := models.NewSensor(ctx,
		models.WithName("Power State"),
		models.WithID("power_state"),
		models.WithState("Powered Off"),
	)

	// The old update fails (server 500) so it is stored pending. The newer
	// update succeeds (server 200) and clears the pending entry.
	entityCh <- old
	entityCh <- newer

	// Wait for the newer update to succeed.
	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return successfulRequests >= 1
	}, 3*time.Second, 20*time.Millisecond)

	// Let the retry ticker fire a few times. The pending old state was cleared
	// by the newer success, so it must not be sent again.
	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	// Exactly one successful request (the newer state), not the old one again.
	assert.Equal(t, 1, successfulRequests)
}
