// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package models

import (
	"fmt"
	"strings"
)

// NewSuccessMessage creates a new Message indicating success with the given summary and (optional) details.
func NewSuccessMessage(summary string, details string) *Message {
	return &Message{
		Status:  MessageStatusSuccess,
		Summary: summary,
		Details: details,
	}
}

// NewErrorMessage creates a new Message indicating an error with the given summary and (optional) details.
func NewErrorMessage(summary string, details string) *Message {
	return &Message{
		Status:  MessageStatusError,
		Summary: summary,
		Details: details,
	}
}

// NewWarningMessage creates a new Message indicating a warning with the given summary and (optional) details.
func NewWarningMessage(summary string, details string) *Message {
	return &Message{
		Status:  MessageStatusWarning,
		Summary: summary,
		Details: details,
	}
}

// NewInfoMessage creates a new Message indicating informational details with the given summary and (optional) details.
func NewInfoMessage(summary string, details string) *Message {
	return &Message{
		Status:  MessageStatusInfo,
		Summary: summary,
		Details: details,
	}
}

// HasDetails returns a boolean indicating whether the message has additional details.
func (msg *Message) HasDetails() bool {
	return msg.Details != ""
}

// String returns the message as a formatted string. This allows Message to satisfy the Stringer interface.
func (msg *Message) String() string {
	var str strings.Builder
	str.WriteString(fmt.Sprintf("%s: %s", strings.ToTitle(string(msg.Status)), msg.Summary))
	if msg.Details != "" {
		str.WriteString("\n" + msg.Details)
	}
	return str.String()
}

// IsSuccess returns true when the message indicates success.
func (msg *Message) IsSuccess() bool {
	return msg.Status == MessageStatusSuccess
}

// IsError returns true when the message indicates an error.
func (msg *Message) IsError() bool {
	return msg.Status == MessageStatusError
}

// IsWarning returns true when the message indicates a warning.
func (msg *Message) IsWarning() bool {
	return msg.Status == MessageStatusWarning
}

// IsInfo returns true when the message indicates informational status.
func (msg *Message) IsInfo() bool {
	return msg.Status == MessageStatusInfo
}
