// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import "os"

const (
	envProcFSRoot = "PROCFS_ROOT"
	envDevFSRoot  = "DEVFS_ROOT"
	envSysFSRoot  = "SYSFS_ROOT"
)

var (
	ProcFSRoot = "/proc"
	DevFSRoot  = "/dev"
	SysFSRoot  = "/sys"
)

func init() {
	var (
		value string
		found bool
	)

	value, found = os.LookupEnv(envProcFSRoot)
	if found {
		ProcFSRoot = value
	}

	value, found = os.LookupEnv(envDevFSRoot)
	if found {
		DevFSRoot = value
	}

	value, found = os.LookupEnv(envSysFSRoot)
	if found {
		SysFSRoot = value
	}
}
