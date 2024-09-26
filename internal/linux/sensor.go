// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"errors"
)

const (
	DataSrcDbus    = "D-Bus"
	DataSrcProcfs  = "ProcFS"
	DataSrcSysfs   = "SysFS"
	DataSrcNetlink = "Netlink"
)

var ErrUnimplemented = errors.New("unimplemented functionality")
