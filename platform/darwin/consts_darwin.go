package darwin

const (
	// DataSrcGetfsstat indicates that the source of this data is from the getfsstat syscall.
	DataSrcGetfsstat = "getfsstat"
	// DataSrcIOKit indicates that the source of this data is from IOKit.
	DataSrcIOKit = "IOKit"
	// DataSrcMach indicates that the source of this data is a Mach kernel host port call (e.g. host_processor_info, host_statistics64).
	DataSrcMach = "Mach"
	// DataSrcSysctl indicates that the source of this data is the sysctl interface.
	DataSrcSysctl = "sysctl"
	// DataSrcIoreg indicates that the source of this data is from the ioreg command.
	DataSrcIoreg = "ioreg"
)
