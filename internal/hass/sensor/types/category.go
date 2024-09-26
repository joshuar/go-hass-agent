// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//go:generate stringer -type=Category -output category_generated.go -linecomment
package types

const (
	CategoryDefault    Category = iota //
	CategoryDiagnostic                 // diagnostic
)

type Category int
