#!/usr/bin/env bash

# Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
# 
# This software is released under the MIT License.
# https://opensource.org/licenses/MIT

# Stop on errors
set -e

sudo apt -y update 

# Install go-hass-agent dependencies
export DEBIAN_FRONTEND=noninteractive && \
    sudo apt -y install libgl1-mesa-dev libxi-dev libxcursor-dev libxrandr-dev libxinerama-dev libxxf86vm-dev dbus-x11 desktop-file-utils fish
cd /workspaces/go-hass-agent && go mod download

# Install goreleaser 
go install github.com/goreleaser/goreleaser@latest

# Install go build packages
go install golang.org/x/tools/cmd/stringer@latest
go install github.com/fyne-io/fyne-cross@latest
go install golang.org/x/text/cmd/gotext@latest
go install github.com/matryer/moq@latest

