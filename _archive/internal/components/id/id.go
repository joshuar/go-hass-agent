// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Package id contains methods for generating universally unique IDs.
package id

//go:generate go tool golang.org/x/tools/cmd/stringer -type=Prefix -linecomment -output id_generated.go

import (
	"fmt"
	"strings"

	nanoid "github.com/matoous/go-nanoid"

	"github.com/joshuar/go-hass-agent/internal/models"
)

const (
	// Unknown represents an unknown prefix.
	Unknown Prefix = iota // unknown
	// ScriptJob prefix is for scheduler jobs for scripts.
	ScriptJob // script_job
	// HassJob prefix is for scheduler jobs for the hass backend.
	HassJob // hass_job
	// Worker prefix is for entity/mqtt workers.
	Worker // worker
)

// Prefix represents a type of ID. Specific types share a common prefix.
type Prefix int

// NewID generates a new unique ID for the given type option. If an ID cannot be
// generated, a non-nil error is returned.
func NewID(option Prefix) (models.ID, error) {
	id, err := nanoid.Nanoid()
	if err != nil {
		return "", fmt.Errorf("could not generate username: %w", err)
	}

	return option.String() + "_" + id, nil
}

// IdentifyID takes an ID and returns the type of ID it represents.
func IdentifyID(id models.ID) Prefix {
	idParts := strings.Split(id, "_")
	switch idParts[0] {
	case HassJob.String():
		return HassJob
	case ScriptJob.String():
		return ScriptJob
	default:
		return Unknown
	}
}
