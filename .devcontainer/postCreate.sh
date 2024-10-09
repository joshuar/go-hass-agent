#!/usr/bin/env bash

set -e

sudo ./build/scripts/enable-multiarch all
sudo ./build/scripts/install-build-deps all ubuntu

# Install and configure starship
curl -sS https://starship.rs/install.sh | sh -s -- -y || exit -1
mkdir -p ~/.config/fish
echo "starship init fish | source" >>~/.config/fish/config.fish
exit 0
