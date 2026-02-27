// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package pipewire

import (
	"errors"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
)

const defaultSink = "@DEFAULT_AUDIO_SINK@"

// runCmd runs a command and returns its trimmed stdout output.
func runCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", fmt.Errorf("%s failed: %s", name, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", fmt.Errorf("%s: %w", name, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetVolume returns the current volume of the default sink as a value 0.0–1.0.
func GetVolume() (float64, error) {
	out, err := runCmd("wpctl", "get-volume", defaultSink)
	if err != nil {
		return 0, err
	}
	// Output format: "Volume: 0.75" or "Volume: 0.75 [MUTED]"
	fields := strings.Fields(out)
	if len(fields) < 2 {
		return 0, fmt.Errorf("unexpected wpctl output: %q", out)
	}
	vol, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse volume from %q: %w", out, err)
	}
	return vol, nil
}

// IsMuted returns true if the default sink is muted.
func IsMuted() (bool, error) {
	out, err := runCmd("wpctl", "get-volume", defaultSink)
	if err != nil {
		return false, err
	}
	return strings.Contains(out, "[MUTED]"), nil
}

// SetVolume sets the volume of the default sink. vol is clamped to [0.0, 1.5].
// Values above 1.0 boost beyond 100% (use with care).
func SetVolume(vol float64) error {
	vol = math.Max(0, math.Min(1.5, vol))
	volStr := fmt.Sprintf("%.2f", vol)
	_, err := runCmd("wpctl", "set-volume", defaultSink, volStr)
	return err
}

// ChangeVolume increases or decreases volume by a percentage amount (e.g. 5 means +5%).
func ChangeVolume(deltaPercent float64) error {
	current, err := GetVolume()
	if err != nil {
		return err
	}
	return SetVolume(current + deltaPercent/100.0)
}

// Mute mutes the default sink.
func Mute() error {
	_, err := runCmd("wpctl", "set-mute", defaultSink, "1")
	return err
}

// Unmute unmutes the default sink.
func Unmute() error {
	_, err := runCmd("wpctl", "set-mute", defaultSink, "0")
	return err
}

// ToggleMute toggles the mute state of the default sink.
func ToggleMute() error {
	_, err := runCmd("wpctl", "set-mute", defaultSink, "toggle")
	return err
}

// Status prints a human-readable status of the default audio sink.
func Status() error {
	out, err := runCmd("wpctl", "status")
	if err != nil {
		return err
	}

	// Print only the Audio section for readability
	lines := strings.Split(out, "\n")
	inAudio := false
	for _, line := range lines {
		if strings.Contains(line, "Audio") {
			inAudio = true
		} else if inAudio && len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
			// New top-level section started
			break
		}
		if inAudio {
			fmt.Println(line)
		}
	}
	return nil
}
