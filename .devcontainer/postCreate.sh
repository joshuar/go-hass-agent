#!/usr/bin/env bash

set -e

cd /workspaces/go-hass-agent && go mod tidy

# Install go build packages
go install golang.org/x/tools/cmd/stringer@latest
go install github.com/fyne-io/fyne-cross@latest
go install golang.org/x/text/cmd/gotext@latest
go install github.com/matryer/moq@latest
go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
go install github.com/tomwright/dasel/v2/cmd/dasel@latest
go install github.com/sigstore/cosign/v2/cmd/cosign@latest

exit 0