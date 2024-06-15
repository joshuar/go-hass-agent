#!/usr/bin/env bash

set -e

# Install go build packages
go install golang.org/x/tools/cmd/stringer@latest
go install golang.org/x/text/cmd/gotext@latest
go install github.com/matryer/moq@latest
go install github.com/tomwright/dasel/v2/cmd/dasel@latest
go install github.com/sigstore/cosign/v2/cmd/cosign@latest

exit 0