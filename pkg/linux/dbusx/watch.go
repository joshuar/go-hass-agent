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

// MatchPath matches messages which are sent from or to the given object. An
// example of a path match is path='/org/freedesktop/Hal/Manager'
//
// https://dbus.freedesktop.org/doc/dbus-specification.html#message-protocol-marshaling-object-path
func MatchPath(path string) WatchOption {
	return func(a *Watch) {
		a.path = dbus.ObjectPath(path)
		a.matches = append(a.matches, dbus.WithMatchObjectPath(a.path))
	}
}

// MatchPathNamespace matches messages which are sent from or to an object for
// which the object path is either the given value, or that value followed by
// one or more path components. For valid paths, see:
//
// https://dbus.freedesktop.org/doc/dbus-specification.html#message-protocol-marshaling-object-path
func MatchPathNamespace(path string) WatchOption {
	return func(a *Watch) {
		a.pathNamespace = path
		a.matches = append(a.matches, dbus.WithMatchPathNamespace(dbus.ObjectPath(path)))
	}
}

// MatchInterface match messages sent over or to a particular interface. An
// example of an interface match is interface='org.freedesktop.Hal.Manager'. For
// valid interfaces, see:
//
// https://dbus.freedesktop.org/doc/dbus-specification.html#message-protocol-names-interface
func MatchInterface(intr string) WatchOption {
	return func(a *Watch) {
		a.matches = append(a.matches, dbus.WithMatchInterface(intr))
	}
}

// MatchMembers matches messages which have the give method or signal name. An
// example of a member match is member='NameOwnerChanged'. Each member match
// will generate a separate watch automatically.
func MatchMembers(names ...string) WatchOption {
	return func(a *Watch) {
		a.methods = append(a.methods, names...)
	}
}

// MatchArgs matches are special and are used for further restricting the match
// based on the arguments in the body of a message. An example of an argument
// match would be arg3='Foo'. Only argument indexes from 0 to 63 should be
// accepted.
func MatchArgs(args map[int]string) WatchOption {
	return func(a *Watch) {
		for arg, value := range args {
			a.matches = append(a.matches, dbus.WithMatchArg(arg, value))
		}
	}
}

// MatchArgNameSpace matches messages whose first argument is the given type,
// and is a bus name or interface name within the specified namespace. This is
// primarily intended for watching name owner changes for a group of related bus
// names, rather than for a single name or all name changes.
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

// NewWatch will create a new D-Bus watch with the given options.
func NewWatch(options ...WatchOption) *Watch {
	watch := &Watch{}
	for _, option := range options {
		option(watch)
	}

	return watch
}

// Start will set up a channel on which D-Bus messages matching the given
// rules can be monitored. Typically, this is used to react when a certain
// property or signal with a given path and on a given interface, changes. The
// data returned in the channel will contain the signal (or property) that
// triggered the match, the path and the contents (what values actually
// changed).
//
//nolint:gocognit
func (w *Watch) Start(ctx context.Context, bus *Bus) (chan Trigger, error) {
	if len(w.methods) > 0 { // Set up a watch for on each method plus all other conditions specified.
		for _, method := range w.methods {
			matches := append(w.matches, dbus.WithMatchMember(method))
			if err := bus.conn.AddMatchSignalContext(ctx, matches...); err != nil {
				return nil, fmt.Errorf("unable to add watch conditions (%v): %w", w.matches, err)
			}

			go func() {
				<-ctx.Done()

				if err := bus.conn.RemoveMatchSignal(matches...); err != nil {
					bus.traceLog("Unable to remove match signal.",
						slog.Any("matches", matches),
						slog.Any("error", err))
				}
			}()

			bus.traceLog("Added D-Bus watch.", slog.Any("matches", matches))
		}
	} else { // Set up a watch on the specified conditions.
		if err := bus.conn.AddMatchSignalContext(ctx, w.matches...); err != nil {
			return nil, fmt.Errorf("unable to add watch conditions (%v): %w", w.matches, err)
		}

		go func() {
			<-ctx.Done()

			if err := bus.conn.RemoveMatchSignal(w.matches...); err != nil {
				bus.traceLog("Unable to remove match signal.",
					slog.Any("matches", w.matches),
					slog.Any("error", err))
			}
		}()

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
	// signals and data. If the context is canceled (i.e., agent shutdown),
	// clean up.
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(outCh)
				bus.conn.RemoveSignal(signalCh)

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
