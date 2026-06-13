// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package webcam

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"strconv"

	slogctx "github.com/veqryn/slog-context"
)

// inputArgs returns the platform-specific ffmpeg input flags.
func inputArgs(device string, fps, width, height int) []string {
	switch runtime.GOOS {
	case "darwin":
		// AVFoundation: device index like "0"
		return []string{
			"-f", "avfoundation",
			"-framerate", strconv.Itoa(fps),
			"-video_size", fmt.Sprintf("%dx%d", width, height),
			"-i", device + ":none", // video only, no audio
		}
	case "windows":
		// DirectShow: device is a friendly name or index like "video=0"
		name := device
		if _, err := strconv.Atoi(device); err == nil {
			name = "video=" + device
		}
		return []string{
			"-f", "dshow",
			"-framerate", strconv.Itoa(fps),
			"-video_size", fmt.Sprintf("%dx%d", width, height),
			"-i", name,
		}
	default: // Linux / V4L2
		dev := device
		if _, err := strconv.Atoi(device); err == nil {
			dev = "/dev/video" + device
		}
		return []string{
			"-f", "v4l2",
			"-framerate", strconv.Itoa(fps),
			"-video_size", fmt.Sprintf("%dx%d", width, height),
			"-i", dev,
		}
	}
}

// capture runs ffmpeg and feeds JPEG frames into h.
// It restarts automatically on crash (up to maxRetries in a row).
func Capture(ctx context.Context, ffmpegBin, device string, fps, width, height int, frameCh chan []byte) {
	const maxRetries = 5
	retries := 0

	for {
		select {
		case <-ctx.Done():
			close(frameCh)
			return
		default:
			slogctx.FromCtx(ctx).Debug("Starting ffmpeg capture",
				slog.String("device", device),
				slog.Int("width", width),
				slog.Int("height", height),
				slog.Int("fps", fps))

			args := inputArgs(device, fps, width, height)
			// Output: raw MJPEG stream to stdout
			args = append(args,
				"-vf", fmt.Sprintf("fps=%d", fps),
				"-f", "mjpeg",
				"-q:v", "3", // JPEG quality 1–31, lower = better
				"pipe:1",
			)

			cmd := exec.CommandContext(ctx, ffmpegBin, args...)
			cmd.Stderr = os.Stderr // ffmpeg logs → our stderr
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				slogctx.FromCtx(ctx).Warn("StdoutPipe failed.",
					slog.Any("error", err))
				return
			}
			if err := cmd.Start(); err != nil {
				slogctx.FromCtx(ctx).Warn("Failed to start ffmpeg.",
					slog.Any("error", err))
				return
			}

			if err := parseMJPEG(stdout, frameCh); err != nil && err != io.EOF {
				slogctx.FromCtx(ctx).Warn("Failed to parse mjpeg.",
					slog.Any("error", err))
				return
			}

			_ = cmd.Wait()
			retries++
			if retries >= maxRetries {
				slogctx.FromCtx(ctx).Warn("ffmpeg crashed!")
				return
			}
		}
	}
}

var (
	jpegSOI = []byte{0xFF, 0xD8}
	jpegEOI = []byte{0xFF, 0xD9}
)

func parseMJPEG(r io.Reader, frameCh chan []byte) error {
	buf := make([]byte, 0, 1<<20) // 1 MB initial cap
	tmp := make([]byte, 32*1024)

	for {
		n, err := r.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
			buf = emitFrames(buf, frameCh)
		}
		if err != nil {
			return fmt.Errorf("read frame: %w", err)
		}
	}
}

// emitFrames extracts complete JPEG frames from buf, publishes them, and
// returns the unconsumed tail.
func emitFrames(buf []byte, frameCh chan []byte) []byte {
	for {
		start := bytes.Index(buf, jpegSOI)
		if start == -1 {
			return buf[:0] // no SOI yet
		}
		// Look for EOI after SOI
		end := bytes.Index(buf[start+2:], jpegEOI)
		if end == -1 {
			// Frame not complete; keep from SOI onward
			return buf[start:]
		}
		end = start + 2 + end + 2 // absolute index, inclusive of EOI
		frame := buf[start:end]
		cp := make([]byte, len(frame))
		copy(cp, frame)
		frameCh <- cp
		buf = buf[end:]
	}
}
