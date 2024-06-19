#!/usr/bin/env bash

set -e

# Install libraries for all supported arches
sudo ./build/scripts/enable-multiarch all
sudo ./build/scripts/install-deps arm arm64 amd64 

# Install go build packages
go install golang.org/x/tools/cmd/stringer@latest
go install golang.org/x/text/cmd/gotext@latest
go install github.com/matryer/moq@latest
go install github.com/tomwright/dasel/v2/cmd/dasel@latest
go install github.com/sigstore/cosign/v2/cmd/cosign@latest
go install github.com/magefile/mage@latest

exit 0