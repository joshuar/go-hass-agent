#!/usr/bin/env bash

set -e

# Install additional packages.
sudo apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && sudo apt-get -y install --no-install-recommends micro pre-commit graphviz \
    && sudo apt-get -y autoremove && sudo apt-get -y clean && sudo rm -rf /var/lib/apt/lists/*

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

# Install build dependencies for all supported arch.
# sudo ./build/scripts/enable-multiarch all
# sudo ./build/scripts/install-build-deps all ubuntu
# sudo ./build/scripts/install-run-deps linux/amd64

exit 0
