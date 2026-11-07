// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package mqtt

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/models"
)

// recordingPublisher records the batches of messages it is asked to publish.
type recordingPublisher struct {
	mu      sync.Mutex
	batches [][]*models.MQTTMsg
	err     error
}

func (p *recordingPublisher) publish(_ context.Context, msgs ...*models.MQTTMsg) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.batches = append(p.batches, msgs)

	return p.err
}

func (p *recordingPublisher) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	return len(p.batches)
}

func TestRepublishStates(t *testing.T) {
	states := func() []*models.MQTTMsg {
		return []*models.MQTTMsg{
			mqttapi.NewMsg("test/state/one", []byte("1")),
			mqttapi.NewMsg("test/state/two", []byte("2")),
		}
	}
	const delay = 50 * time.Millisecond

	t.Run("publishes states after online signal and delay", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		haOnline := make(chan struct{}, 1)
		pub := &recordingPublisher{}

		go republishStates(ctx, haOnline, delay, states, pub.publish)

		haOnline <- struct{}{}
		require.Never(t, func() bool { return pub.count() > 0 }, delay/2, time.Millisecond,
			"states published before delay elapsed")
		require.Eventually(t, func() bool { return pub.count() == 1 }, time.Second, time.Millisecond)

		pub.mu.Lock()
		require.Len(t, pub.batches[0], 2)
		require.Equal(t, "test/state/one", pub.batches[0][0].Topic)
		pub.mu.Unlock()
	})

	t.Run("coalesces online signals received while waiting", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		haOnline := make(chan struct{}, 1)
		pub := &recordingPublisher{}

		go republishStates(ctx, haOnline, delay, states, pub.publish)

		haOnline <- struct{}{}
		// Wait until the first signal has been consumed, then send another
		// while the delay is pending.
		require.Eventually(t, func() bool { return len(haOnline) == 0 }, time.Second, time.Millisecond)
		haOnline <- struct{}{}

		require.Eventually(t, func() bool { return pub.count() == 1 }, time.Second, time.Millisecond)
		require.Never(t, func() bool { return pub.count() > 1 }, 2*delay, time.Millisecond)
	})

	t.Run("republishes on each separate online signal", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		haOnline := make(chan struct{}, 1)
		pub := &recordingPublisher{err: errors.New("publish failed")}

		go republishStates(ctx, haOnline, delay, states, pub.publish)

		haOnline <- struct{}{}
		require.Eventually(t, func() bool { return pub.count() == 1 }, time.Second, time.Millisecond)
		// A failed publish must not stop later republishes.
		haOnline <- struct{}{}
		require.Eventually(t, func() bool { return pub.count() == 2 }, time.Second, time.Millisecond)
	})

	t.Run("does not publish when cancelled during delay", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		haOnline := make(chan struct{}, 1)
		pub := &recordingPublisher{}
		done := make(chan struct{})

		go func() {
			republishStates(ctx, haOnline, delay, states, pub.publish)
			close(done)
		}()

		haOnline <- struct{}{}
		require.Eventually(t, func() bool { return len(haOnline) == 0 }, time.Second, time.Millisecond)
		cancel()

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("republishStates did not return after cancellation")
		}
		require.Equal(t, 0, pub.count())
	})
}
