#!/usr/bin/env bash

set -e

sudo ./build/scripts/enable-multiarch all
sudo ./build/scripts/install-build-deps all ubuntu

# Install go build packages
go install github.com/magefile/mage@9e91a03eaa438d0d077aca5654c7757141536a60                 # v1.15.0
go install github.com/sigstore/cosign/v2/cmd/cosign@b5e7dc123a272080f4af4554054797296271e902 # v2.4.0

# Install and configure starship
curl -sS https://starship.rs/install.sh | sh -s -- -y || exit -1
mkdir -p ~/.config/fish
echo "starship init fish | source" >>~/.config/fish/config.fish
exit 0
