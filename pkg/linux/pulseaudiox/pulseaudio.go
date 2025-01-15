// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package pulseaudiox

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/jfreymuth/pulse/proto"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
)

const (
	volumeMaxPc = 100
	volumeMinPc = 0
)

var ErrVolumeOutofRange = errors.New("volume out of range")

// PulseAudioClient represents a connection to PulseAudio. It will have an event
// channel which is triggered any time a change happens on the default output
// device. It also records the current state of the volume and mute status.
type PulseAudioClient struct {
	client  *proto.Client
	conn    net.Conn
	EventCh chan struct{}
	doneCh  chan struct{}
	// Mute (true: muted, false: unmuted).
	Mute bool
	// Vol as a percent (0 - 100%).
	Vol float64
}

// Default input/output devices.
const (
	outputDevice = "@DEFAULT_SINK@"
	inputDevice  = "@DEFAULT_SOURCE@"
)

// NewPulseClient creates a new connection to Pulseaudio. It also retrieves the
// current mute state and volume level of the default output device. It will
// also set up an event channel that can be used to listen to when a change is
// made to the output device (volume changed, mute status changed, etc.) If it
// cannot connect to Pulseaudio, a non-nil error will be returned with details
// on the issue.
//
//nolint:cyclop
//revive:disable:unnecessary-stmt
func NewPulseClient(ctx context.Context) (*PulseAudioClient, error) {
	// Connect to pulseaudio.
	client, conn, err := proto.Connect("")
	if err != nil {
		return nil, fmt.Errorf("could not connect to pulseaudio: %w", err)
	}

	pulse := &PulseAudioClient{
		client:  client,
		conn:    conn,
		EventCh: make(chan struct{}, 1),
		doneCh:  make(chan struct{}, 1),
	}
	// Set client properties.
	props := proto.PropList{
		"media.name":                 proto.PropListString(preferences.AppName),
		"application.name":           proto.PropListString("go-hass-agent"),
		"application.icon_name":      proto.PropListString("audio-x-generic"),
		"application.process.id":     proto.PropListString(strconv.Itoa(os.Getpid())),
		"application.process.binary": proto.PropListString("go-hass-agent"),
		"window.x11.display":         proto.PropListString(os.Getenv("DISPLAY")),
	}

	err = pulse.client.Request(&proto.SetClientName{Props: props}, &proto.SetClientNameReply{})
	if err != nil {
		return nil, fmt.Errorf("could not send client info: %w", err)
	}

	// Get current mute state.
	muteState, err := pulse.GetMute()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve current mute status: %w", err)
	}

	pulse.Mute = muteState

	// Get current volume.
	volPct, err := pulse.GetVolume()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve current volume: %w", err)
	}

	pulse.Vol = volPct

	// Callback function to be used when a Pulseaudio event occurs.
	pulse.client.Callback = func(val any) {
		switch val := val.(type) {
		case *proto.SubscribeEvent:
			if val.Event.GetType() == proto.EventChange && val.Event.GetFacility() == proto.EventSink {
				select {
				case <-pulse.doneCh:
					close(pulse.EventCh)

					return
				case pulse.EventCh <- struct{}{}:
				default:
				}
			}
		}
	}

	// Request to subscribe to all events.
	err = pulse.client.Request(&proto.Subscribe{Mask: proto.SubscriptionMaskAll}, nil)
	if err != nil {
		return nil, fmt.Errorf("could not subscribe to pulseaudio events: %w", err)
	}

	// Shutdown gracefully when requested.
	go func() {
		defer pulse.conn.Close()
		defer close(pulse.doneCh)
		<-ctx.Done()
	}()

	return pulse, nil
}

// GetVolume will retrieve the current volume of the default output device, as a
// percentage.
func (c *PulseAudioClient) GetVolume() (float64, error) {
	repl, err := c.GetState()
	if err != nil {
		return -1, fmt.Errorf("could not get current state: %w", err)
	}

	volPct := ParseVolume(repl)

	return volPct, nil
}

// SetVolume will set the volume of the default output device to the given
// percent amount. Values outside of 0 - 100 will be rejected.

func (c *PulseAudioClient) SetVolume(pct float64) error {
	if pct < volumeMinPc || pct > volumeMaxPc {
		return ErrVolumeOutofRange
	}

	repl, err := c.GetState()
	if err != nil {
		return fmt.Errorf("could not set volume: %w", err)
	}

	newVolume := pct / volumeMaxPc * float64(proto.VolumeNorm)
	volumes := repl.ChannelVolumes

	for i := range volumes {
		volumes[i] = uint32(newVolume)
	}

	err = c.client.Request(&proto.SetSinkVolume{SinkIndex: proto.Undefined, SinkName: outputDevice, ChannelVolumes: volumes}, nil)
	if err != nil {
		return fmt.Errorf("could not set volume: %w", err)
	}

	c.Vol = pct

	return nil
}

// GetMute retrieve the current mute state of the default output device as a
// bool (true: muted, false: unmuted).
func (c *PulseAudioClient) GetMute() (bool, error) {
	repl, err := c.GetState()
	if err != nil {
		return false, fmt.Errorf("could not get current state: %w", err)
	}

	return repl.Mute, nil
}

// SetMute will set the mute state of the default output device to the given
// state.
func (c *PulseAudioClient) SetMute(state bool) error {
	err := c.client.Request(&proto.SetSinkMute{SinkIndex: proto.Undefined, SinkName: outputDevice, Mute: state}, nil)
	if err != nil {
		return fmt.Errorf("could not set mute state: %w", err)
	}

	c.Mute = state

	return nil
}

// GetState will return the low-level current state representation of the
// default output device. It can be used for more advanced parsing and retrieval
// about the output device.
func (c *PulseAudioClient) GetState() (*proto.GetSinkInfoReply, error) {
	repl := &proto.GetSinkInfoReply{}

	err := c.client.Request(&proto.GetSinkInfo{SinkIndex: proto.Undefined, SinkName: outputDevice}, repl)
	if err != nil {
		return nil, fmt.Errorf("could not parse reply: %w", err)
	}

	return repl, nil
}

// ParseVolume will retrieve the volume as a percentage from a state message.
func ParseVolume(repl *proto.GetSinkInfoReply) float64 {
	var acc int64
	for _, vol := range repl.ChannelVolumes {
		acc += int64(vol)
	}

	acc /= int64(len(repl.ChannelVolumes))

	return float64(acc) / float64(proto.VolumeNorm) * volumeMaxPc
}
