// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package pipewire

import (
	"slices"
	"sync"
)

// NodeTracker tracks the state of every node of a given media class.
//
// A system usually has several nodes of the same class (e.g. a built-in
// microphone, a webcam microphone and a Bluetooth headset). Deriving "in use"
// from the most recent event alone is wrong: an unrelated node changing to
// suspended would mark the class as not in use while another node is still
// running. NodeTracker instead keeps the last known state per node id.
//
// Use NodeTracker.Filter as the filter function passed to Monitor.AddListener.
// The filter runs synchronously in the monitor's decode loop, so the tracked
// state follows pw-dump's event order even though events are delivered to
// listeners asynchronously.
type NodeTracker struct {
	class MediaClass
	mu    sync.Mutex
	nodes map[int]State
}

// NewNodeTracker creates a tracker for nodes of the given media class.
func NewNodeTracker(class MediaClass) *NodeTracker {
	return &NodeTracker{
		class: class,
		nodes: make(map[int]State),
	}
}

// Filter records the state of nodes of the tracked media class. It returns
// true for any event that changed the tracked set or a tracked node's state.
func (t *NodeTracker) Filter(event *Event) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if event.IsRemovalEvent() {
		return t.remove(event.ID)
	}

	if event.Type != InterfaceNodeEvent || event.Info == nil {
		return false
	}

	prev, tracked := t.nodes[event.ID]
	if !t.isTrackedClass(event, tracked) {
		return t.remove(event.ID)
	}

	state := prev
	if event.Info.State != nil {
		state = *event.Info.State
	}

	t.nodes[event.ID] = state

	return !tracked || state != prev
}

// isTrackedClass reports whether the event's node has the tracked media class.
// Props are not guaranteed to be present on every event, so without them it
// falls back to whether the node is already tracked.
func (t *NodeTracker) isTrackedClass(event *Event, tracked bool) bool {
	if event.Info.Props == nil {
		return tracked
	}

	props, err := event.NodeProps()

	return err == nil && props.MediaClass == t.class
}

// remove stops tracking a node, reporting whether it was tracked.
func (t *NodeTracker) remove(id int) bool {
	_, found := t.nodes[id]
	delete(t.nodes, id)

	return found
}

// AnyIn reports whether any tracked node is in one of the given states.
func (t *NodeTracker) AnyIn(states ...State) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, state := range t.nodes {
		if slices.Contains(states, state) {
			return true
		}
	}

	return false
}
