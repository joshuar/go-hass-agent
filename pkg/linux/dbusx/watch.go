// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package dbusx

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/godbus/dbus/v5"
)

type Trigger struct {
	Signal  string
	Path    string
	Content []any
}

type Watch struct {
	path          dbus.ObjectPath
	pathNamespace string
	matches       []dbus.MatchOption
	methods       []string
}

type WatchOption func(*Watch)

func MatchPath(path string) WatchOption {
	return func(a *Watch) {
		a.path = dbus.ObjectPath(path)
		a.matches = append(a.matches, dbus.WithMatchObjectPath(a.path))
	}
}

func MatchPathNamespace(path string) WatchOption {
	return func(a *Watch) {
		a.pathNamespace = path
		a.matches = append(a.matches, dbus.WithMatchPathNamespace(dbus.ObjectPath(path)))
	}
}

func MatchInterface(intr string) WatchOption {
	return func(a *Watch) {
		a.matches = append(a.matches, dbus.WithMatchInterface(intr))
	}
}

func MatchMembers(names ...string) WatchOption {
	return func(a *Watch) {
		a.methods = append(a.methods, names...)
	}
}

func MatchArgs(args map[int]string) WatchOption {
	return func(a *Watch) {
		for arg, value := range args {
			a.matches = append(a.matches, dbus.WithMatchArg(arg, value))
		}
	}
}

func MatchArgNameSpace(name string) WatchOption {
	return func(a *Watch) {
		a.matches = append(a.matches, dbus.WithMatchArg0Namespace(name))
	}
}

// MatchPropChanged will set up a D-Bus watch to match on the
// org.freedesktop.DBus.Properties.PropertiesChanged signal of the
// org.freedesktop.DBus.Properties interface, together with other match options.
func MatchPropChanged() WatchOption {
	return func(w *Watch) {
		w.matches = append(w.matches,
			dbus.WithMatchInterface(PropInterface),
			dbus.WithMatchMember("PropertiesChanged"),
		)
	}
}

func NewWatch(options ...WatchOption) *Watch {
	watch := &Watch{}
	for _, option := range options {
		option(watch)
	}

	return watch
}

// WatchBus will set up a channel on which D-Bus messages matching the given
// rules can be monitored. Typically, this is used to react when a certain
// property or signal with a given path and on a given interface, changes. The
// data returned in the channel will contain the signal (or property) that
// triggered the match, the path and the contents (what values actually
// changed).
//
//nolint:cyclop,gocognit
func (w *Watch) Start(ctx context.Context, bus *Bus) (chan Trigger, error) {
	if len(w.methods) > 0 { // Set up a watch for on each method plus all other conditions specified.
		for _, method := range w.methods {
			matches := append(w.matches, dbus.WithMatchMember(method))
			if err := bus.conn.AddMatchSignalContext(ctx, matches...); err != nil {
				return nil, fmt.Errorf("unable to add watch conditions (%v): %w", w.matches, err)
			}

			bus.traceLog("Added D-Bus watch.", slog.Any("matches", matches))
		}
	} else { // Set up a watch on the specified conditions.
		if err := bus.conn.AddMatchSignalContext(ctx, w.matches...); err != nil {
			return nil, fmt.Errorf("unable to add watch conditions (%v): %w", w.matches, err)
		}

		bus.traceLog("Added D-Bus watch.", slog.Any("matches", w.matches))
	}

	// Set up our channels: signalCh for signals received from D-Bus and outCh
	// where we send the signal.
	signalCh := make(chan *dbus.Signal)
	outCh := make(chan Trigger)
	// Connect our signal chan to the bus signal channel.
	bus.conn.Signal(signalCh)

	// Set up a goroutine to listen for signals from D-Bus and forward them over
	// outCh. We do some generic filtering of the signal to catch obvious bogus
	// signals and data. If the context is cancelled (i.e., agent shutdown),
	// clean up.
	go func() {
		for {
			select {
			case <-ctx.Done():
				bus.conn.RemoveSignal(signalCh)
				close(outCh)

				return
			case signal := <-signalCh:
				// If the signal is empty, ignore.
				if signal == nil {
					continue
				}
				// If a path match was specified and the path in the signal
				// doesn't match it, ignore.
				if w.path != "" {
					if signal.Path != w.path {
						bus.traceLog("Ignoring mismatched path.", slog.Any("signal", signal.Path), slog.Any("match", w.path))

						continue
					}
				}
				// If a path namespace match was specified and the path in the
				// signal is not on that namespace, ignore.
				if w.pathNamespace != "" {
					if !strings.HasPrefix(string(signal.Path), w.pathNamespace) {
						bus.traceLog("Ignoring mismatched path namespace.", slog.Any("signal", signal.Path), slog.Any("match", w.pathNamespace))

						continue
					}
				}
				// We have a match! Send the signal details back to the client
				// for further processing.
				bus.traceLog("Dispatching D-Bus trigger.", slog.Any("signal", signal))

				outCh <- Trigger{
					Signal:  signal.Name,
					Path:    string(signal.Path),
					Content: signal.Body,
				}
			}
		}
	}()

	return outCh, nil
}
