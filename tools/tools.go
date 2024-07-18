//go:build tools

// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tools

import (
	_ "github.com/goreleaser/nfpm/v2/cmd/nfpm"
	_ "github.com/magefile/mage"
	_ "github.com/matryer/moq"
	_ "golang.org/x/text/cmd/gotext"
	_ "golang.org/x/tools/cmd/stringer"
)
