// Code adapted from github.com/ConnorsApps/pipewire-monitor-go.

package pipewire

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type EventType string

const (
	EmptyEvent             EventType = ""
	InterfaceNodeEvent     EventType = "PipeWire:Interface:Node"
	InterfaceMetadataEvent EventType = "PipeWire:Interface:Metadata"
)

type (
	Event struct {
		ID          int             `json:"id"`
		Type        EventType       `json:"type"`
		Version     int             `json:"version"`
		Info        *EventInfo      `json:"info"`
		Permissions []string        `json:"permissions"`
		Metadata    []EventMetadata `json:"metadata,omitempty"`
		Change      string          `json:"change"` // "added" | "changed" | "removed"
		// When the event was received
		CapturedAt time.Time `json:"-"`
	}

	EventInfo struct {
		Direction  string       `json:"direction,omitempty"`
		ChangeMask []string     `json:"change-mask"`
		Props      *EventProps  `json:"props,omitempty"`
		Params     *EventParams `json:"params,omitempty"`
		State      *State       `json:"state,omitempty"`
		Error      *any         `json:"error,omitempty"`
	}

	EventProps struct {
		MediaClass string     `json:"media.class"`
		NodeName   string     `json:"node.name"`
		NodeNick   FlexString `json:"node.nick"`
		NodeDesc   string     `json:"node.description"`
	}

	EventMetadata struct {
		Key   string          `json:"key"`
		Value json.RawMessage `json:"value"` // string, object {"name":"..."}, or null
	}

	EventParams struct {
		EnumFormat []ParamEnumFormat `json:"EnumFormat,omitempty"`
		Meta       []ParamMeta       `json:"Meta,omitempty"`
		IO         []ParamIO         `json:"IO,omitempty"`
		Format     []any             `json:"Format,omitempty"`
		Buffers    []any             `json:"Buffers,omitempty"`
		Latency    []ParamLatency    `json:"Latency,omitempty"`
		Tag        []any             `json:"Tag,omitempty"`
		Props      []ParamProps      `json:"Props,omitempty"`
	}

	ParamEnumFormat struct {
		MediaType    string `json:"mediaType"`
		MediaSubtype string `json:"mediaSubtype"`
		Format       any    `json:"format"`
	}

	ParamMeta struct {
		Type string `json:"type"`
		Size int    `json:"size"`
	}

	ParamIO struct {
		ID   string `json:"id"`
		Size int    `json:"size"`
	}

	ParamProps struct {
		Volume  *float64  `json:"volume"`
		Mute    *bool     `json:"mute"`
		Volumes []float64 `json:"channelVolumes"`
	}

	ParamLatency struct {
		Direction  string  `json:"direction"`
		MinQuantum float64 `json:"minQuantum"`
		MaxQuantum float64 `json:"maxQuantum"`
		MinRate    int     `json:"minRate"`
		MaxRate    int     `json:"maxRate"`
		MinNs      int     `json:"minNs"`
		MaxNs      int     `json:"maxNs"`
	}

	NodeProps struct {
		Name                     string       `json:"node.name"`
		Description              string       `json:"node.description"`
		Nickname                 FlexString   `json:"node.nick"`
		AudioChannels            int          `json:"audio.channels"`
		AudioPosition            string       `json:"audio.position"`
		ClientID                 int          `json:"client.id"`
		DeviceClass              *DeviceClass `json:"device.class"`
		DeviceID                 int          `json:"device.id"`
		DeviceProfileDescription string       `json:"device.profile.description"`
		DeviceProfileName        string       `json:"device.profile.name"`
		FactoryID                int          `json:"factory.id"`
		FactoryMode              string       `json:"factory.mode"`
		FactoryName              string       `json:"factory.name"`
		LibraryName              string       `json:"library.name"`
		MediaClass               MediaClass   `json:"media.class"`
		ObjectID                 int          `json:"object.id"`
		ObjectPath               string       `json:"object.path"`
		ObjectSerial             int          `json:"object.serial"`
	}
)

// FlexString unmarshals a JSON string as-is, but also tolerates non-string
// JSON values (numbers, booleans) by taking their literal representation.
// PipeWire properties are untyped and pw-dump serializes some string
// properties (e.g. "node.nick") as bare JSON numbers when their value looks
// numeric, rather than quoting them.
type FlexString string

func (f *FlexString) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		*f = ""
		return nil
	}

	if trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(trimmed, &s); err != nil {
			return err
		}
		*f = FlexString(s)
		return nil
	}

	*f = FlexString(trimmed)
	return nil
}

type DeviceClass string

const DeviceSound DeviceClass = "sound"

type State string

const (
	StateSuspended State = "suspended"
	StateRunning   State = "running"
	StateIdle      State = "idle"
	StateError     State = "error"
	StateCreating  State = "creating"
)

type MediaClass string

const (
	// MediaAudioSource is a source of audio samples like a microphone.
	MediaAudioSource MediaClass = "Audio/Source"
	// MediaAudioSink is a sink for audio samples, like an audio card.
	MediaAudioSink MediaClass = "Audio/Sink"
	// MediaAudioDuplex is a node that is both a sink and a source.
	MediaAudioDuplex MediaClass = "Audio/Duplex"
	// MediaStreamOutputAudio is a playback stream.
	MediaStreamOutputAudio MediaClass = "Stream/Output/Audio"
	// MediaStreamInputAudio is a capture stream.
	MediaStreamInputAudio MediaClass = "Stream/Input/Audio"
	// MediaVideoSource is a source of video samples like a webcam.
	MediaVideoSource MediaClass = "Video/Source"
)

// IsRemovalEvent indicates the event is an object being removed.
// Example of when an object is removed:
//
//	{
//		"id": 128,
//		"info": null
//	 }
func (e *Event) IsRemovalEvent() bool {
	return e.Info == nil && e.Type == EmptyEvent && e.ID != 0
}

// NodeProps converts the event info to node properties (if possible).
func (e *Event) NodeProps() (*NodeProps, error) {
	if e.Type != InterfaceNodeEvent {
		return nil, errors.New("event is not a node event type")
	} else if e.Info == nil {
		return nil, errors.New("event info is nil")
	}

	var props = &NodeProps{}
	data, err := json.Marshal(e.Info.Props)
	if err != nil {
		return props, fmt.Errorf("marshal node props: %w", err)
	}

	if err = json.Unmarshal(data, props); err != nil {
		return props, fmt.Errorf("unmarshal node props: %w", err)
	}

	return props, nil
}
