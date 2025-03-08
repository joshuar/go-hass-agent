#!/usr/bin/env bash

set -e

# Add starship to fish shell.
mkdir -p ~/.config/fish
echo "starship init fish | source" >>~/.config/fish/config.fish

# Add starship to bash shell.
echo 'eval "$(starship init bash)"' >>~/.bashrc

# Install build dependencies for all supported arch.
sudo ./build/scripts/enable-multiarch all
sudo ./build/scripts/install-build-deps all ubuntu
sudo ./build/scripts/install-run-deps linux/amd64

exit 0
