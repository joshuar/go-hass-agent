#!/usr/bin/bash
# Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
# SPDX-License-Identifier: MIT

set -x

# Add starship to fish shell.
mkdir -p ~/.config/fish
echo "starship init fish | source" >>~/.config/fish/config.fish

# Add starship to bash shell.
echo 'eval "$(starship init bash)"' >>~/.bashrc

cd /workspace

# Update JS packages with bun.
bun update || exit -1

# Install Go tools.
go install github.com/air-verse/air@latest
go install github.com/a-h/templ/cmd/templ@latest

exit 0
