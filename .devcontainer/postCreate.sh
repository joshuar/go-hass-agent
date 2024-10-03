#!/usr/bin/env bash

set -e

sudo ./build/scripts/enable-multiarch all
sudo ./build/scripts/install-build-deps all ubuntu

# Install go build packages
go install golang.org/x/tools/cmd/stringer@v0.23.0
go install golang.org/x/text/cmd/gotext@v0.16.0
go install github.com/matryer/moq@0bf2e8a069abaefdfd07e4902d204441cca17298
go install github.com/magefile/mage@9e91a03eaa438d0d077aca5654c7757141536a60
go install github.com/goreleaser/nfpm/v2/cmd/nfpm@d33a9233bb7acf04621b78114114476196d79779
go install github.com/sigstore/cosign/v2/cmd/cosign@fb651b4ddd8176bd81756fca2d988dd8611f514d

# Install and configure starship
curl -sS https://starship.rs/install.sh | sh -s -- -y || exit -1
mkdir -p ~/.config/fish
echo "starship init fish | source" >>~/.config/fish/config.fish
exit 0
