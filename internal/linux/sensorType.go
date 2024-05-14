// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

//go:generate stringer -type=SensorTypeValue -output sensorTypeStrings.go -linecomment
const (
	SensorUnknown           SensorTypeValue = iota // Unknown Sensor
	SensorAppActive                                // Active App
	SensorAppRunning                               // Running Apps
	SensorBattType                                 // Battery Type
	SensorBattPercentage                           // Battery Level
	SensorBattTemp                                 // Battery Temperature
	SensorBattVoltage                              // Battery Voltage
	SensorBattEnergy                               // Battery Energy
	SensorBattEnergyRate                           // Battery Power
	SensorBattState                                // Battery State
	SensorBattNativePath                           // Battery Path
	SensorBattLevel                                // Battery Level
	SensorBattModel                                // Battery Model
	SensorMemTotal                                 // Memory Total
	SensorMemAvail                                 // Memory Available
	SensorMemUsed                                  // Memory Used
	SensorMemPc                                    // Memory Usage
	SensorSwapTotal                                // Swap Memory Total
	SensorSwapUsed                                 // Swap Memory Used
	SensorSwapFree                                 // Swap Memory Free
	SensorSwapPc                                   // Swap Usage
	SensorConnectionState                          // Connection State
	SensorConnectionID                             // Connection ID
	SensorConnectionDevices                        // Connection Device
	SensorConnectionType                           // Connection Type
	SensorConnectionIPv4                           // Connection IPv4
	SensorConnectionIPv6                           // Connection IPv6
	SensorAddressIPv4                              // IPv4 Address
	SensorAddressIPv6                              // IPv6 Address
	SensorWifiSSID                                 // Wi-Fi SSID
	SensorWifiFrequency                            // Wi-Fi Frequency
	SensorWifiSpeed                                // Wi-Fi Link Speed
	SensorWifiStrength                             // Wi-Fi Signal Strength
	SensorWifiHWAddress                            // Wi-Fi BSSID
	SensorBytesSent                                // Bytes Sent
	SensorBytesRecv                                // Bytes Received
	SensorBytesSentRate                            // Bytes Sent Throughput
	SensorBytesRecvRate                            // Bytes Received Throughput
	SensorPowerProfile                             // Power Profile
	SensorBoottime                                 // Last Reboot
	SensorUptime                                   // Uptime
	SensorLoad1                                    // CPU load average (1 min)
	SensorLoad5                                    // CPU load average (5 min)
	SensorLoad15                                   // CPU load average (15 min)
	SensorCPUPc                                    // CPU Usage
	SensorScreenLock                               // Screen Lock
	SensorLidClosed                                // Lid Closed
	SensorProblem                                  // Problems
	SensorKernel                                   // Kernel Version
	SensorDistribution                             // Distribution Name
	SensorVersion                                  // Distribution Version
	SensorUsers                                    // Current Users
	SensorDeviceTemp                               // Temperature
	SensorPowerState                               // Power State
	SensorAccentColor                              // Accent Color
	SensorColorScheme                              // Color Scheme Type
	SensorDiskReads                                // Disk Reads
	SensorDiskWrites                               // Disk Writes
	SensorDiskReadRate                             // Disk Read Rate
	SensorDiskWriteRate                            // Disk Write Rate
	SensorDocked                                   // Docked State
	SensorExternalPower                            // External Power Connected
)

// SensorTypeValue represents the unique type of sensor data being reported. Every
// sensor will have a different type. A SensorTypeValue maps to an entity in Home
// Assistant.
type SensorTypeValue int
