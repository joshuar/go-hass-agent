// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//go:generate go run golang.org/x/tools/cmd/stringer -type=Prefix -linecomment -output id_generated.go
package id

import (
	"fmt"
	"strings"

	nanoid "github.com/matoous/go-nanoid"
)

const (
	Unknown      Prefix = iota // unknown
	SchedulerJob               // scheduler_job
)

// Prefix represents a type of ID. Specific types share a common prefix.
type Prefix int

// NewID generates a new unique ID for the given type option. If an ID cannot be
// generated, a non-nil error is returned.
func NewID(option Prefix) (string, error) {
	id, err := nanoid.Nanoid()
	if err != nil {
		return "", fmt.Errorf("could not generate username: %w", err)
	}

	return option.String() + "_" + id, nil
}

// IdentifyID takes an ID and returns the type of ID it represents.
func IdentifyID(id string) Prefix {
	idParts := strings.Split(id, "_")
	switch idParts[0] {
	case SchedulerJob.String():
		return SchedulerJob
	default:
		return Unknown
	}
}
