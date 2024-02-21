// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package text

import _ "embed"

//go:embed registerLong.txt
var RegisterCmdLongText string

//go:embed rootLong.txt
var RootCmdLongText string

//go:embed resetLong.txt
var ResetCmdLongText string
