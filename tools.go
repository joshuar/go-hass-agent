//go:build tools
// +build tools

// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

// Package tools imports things required by build scripts and test packages of
// submodules, to force `go mod` to see them as dependencies.
package main

import (
	_ "github.com/davecgh/go-spew/spew"
	_ "github.com/goreleaser/nfpm/v2/cmd/nfpm"
	_ "github.com/magefile/mage"
	_ "github.com/matryer/moq"
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
	_ "github.com/yassinebenaid/godump"
	_ "golang.org/x/tools/cmd/stringer"
)
