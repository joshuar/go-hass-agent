//go:build mage
// +build mage

// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"fmt"

	"github.com/magefile/mage/mg"
)

type Build mg.Namespace

// Full runs all prep steps and then builds the binary.
func (Build) Full(arch string) error {
	fmt.Println("Starting full " + arch + "build...")

	// Make everything nice, neat, and proper
	mg.Deps(Preps.Tidy)
	mg.Deps(Preps.Format)
	mg.Deps(Preps.Generate)

	// Record all licenses in a registry
	mg.Deps(Checks.Licenses)

	return buildProject(arch)
}

// Fast just builds the binary and does not run any prep steps. It will fail if
// the prep steps have not run.
func (Build) Fast(arch string) error {
	fmt.Println("Starting fast " + arch + "build...")

	return buildProject(arch)
}
