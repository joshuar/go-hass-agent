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
export PATH="$HOME/go/bin:/go/bin:/usr/local/go/bin:$PATH" && \
    go mod tidy && \
    go install github.com/air-verse/air@7a8024892f50b53ff20f733de9cbd662a730095b && \
    go install github.com/a-h/templ/cmd/templ@5ddd784440b232930161d76c7ca85d922fdcf183 && \
    go install golang.org/x/tools/gopls@5c4433be420451410e8cfd968eda32a818dac087 && \
    go install github.com/sigstore/cosign/v3/cmd/cosign@479147a4df05f31be48aeb2b3a9d32dfc35ba877 && \
    curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.8.0 && \
    golangci-lint custom && \
    mv /tmp/golangci-lint-v2 $(go env GOPATH)/bin/golangci-lint-v2

# Set up fish shell.
mkdir -p ~/.config/fish
echo 'set --export PATH "/workspace/node_modules/.bin" $PATH' >> ~/.config/fish/config.fish
echo 'set --export PATH "$HOME/go/bin" /go/bin /usr/local/go/bin $PATH' >> ~/.config/fish/config.fish

exit 0
