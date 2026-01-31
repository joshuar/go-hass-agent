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


# Update JS packages.
npm update || exit -1

# Install Go packages.
go mod tidy
go install github.com/air-verse/air@latest
go install github.com/a-h/templ/cmd/templ@latest
go install golang.org/x/tools/gopls@latest
go install github.com/sigstore/cosign/v3/cmd/cosign@latest
curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.8.0
cd /tmp && \
    ~/go/bin/golangci-lint custom && \
    mv golangci-lint $(go env GOPATH)/bin/golangci-lint-v2

# Set up fish shell.
mkdir -p ~/.config/fish
echo 'set --export PATH "/workspace/node_modules/.bin" $PATH' >> ~/.config/fish/config.fish
echo 'set --export PATH "$HOME/go/bin" /go/bin /usr/local/go/bin $PATH' >> ~/.config/fish/config.fish

exit 0
