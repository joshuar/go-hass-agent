// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package linux

import "os"

const (
	// DataSrcDBus indicates that the source of this data is from D-Bus.
	DataSrcDBus = "D-Bus"
	// DataSrcProcFS indicates that the source of this data is from the /proc filesystem.
	DataSrcProcFS = "ProcFS"
	// DataSrcSysFS indicates that the source of this data is from the /sys filesystem.
	DataSrcSysFS = "SysFS"
	// DataSrcNetlink indicates that the source of this data is from Netlink.
	DataSrcNetlink = "Netlink"
)

const (
	envProcFSRoot = "PROCFS_ROOT"
	envDevFSRoot  = "DEVFS_ROOT"
	envSysFSRoot  = "SYSFS_ROOT"
)

var (
	// ProcFSRoot is where the agent expects the /proc filesystem to be mounted.
	ProcFSRoot = "/proc"
	// DevFSRoot is where the agent expects the /dev filesystem to be mounted.
	DevFSRoot = "/dev"
	// SysFSRoot is where the agent expects the /sys filesystem to be mounted.
	SysFSRoot = "/sys"
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
