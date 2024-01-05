// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

//go:generate stringer -type=SensorTypeValue -output sensorTypeStrings.go -linecomment
const (
	SensorAppActive         SensorTypeValue = iota + 1 // Active App
	SensorAppRunning                                   // Running Apps
	SensorBattType                                     // Battery Type
	SensorBattPercentage                               // Battery Level
	SensorBattTemp                                     // Battery Temperature
	SensorBattVoltage                                  // Battery Voltage
	SensorBattEnergy                                   // Battery Energy
	SensorBattEnergyRate                               // Battery Power
	SensorBattState                                    // Battery State
	SensorBattNativePath                               // Battery Path
	SensorBattLevel                                    // Battery Level
	SensorBattModel                                    // Battery Model
	SensorMemTotal                                     // Memory Total
	SensorMemAvail                                     // Memory Available
	SensorMemUsed                                      // Memory Used
	SensorSwapTotal                                    // Swap Memory Total
	SensorSwapUsed                                     // Swap Memory Used
	SensorSwapFree                                     // Swap Memory Free
	SensorConnectionState                              // Connection State
	SensorConnectionID                                 // Connection ID
	SensorConnectionDevices                            // Connection Device
	SensorConnectionType                               // Connection Type
	SensorConnectionIPv4                               // Connection IPv4
	SensorConnectionIPv6                               // Connection IPv6
	SensorAddressIPv4                                  // IPv4 Address
	SensorAddressIPv6                                  // IPv6 Address
	SensorWifiSSID                                     // Wi-Fi SSID
	SensorWifiFrequency                                // Wi-Fi Frequency
	SensorWifiSpeed                                    // Wi-Fi Link Speed
	SensorWifiStrength                                 // Wi-Fi Signal Strength
	SensorWifiHWAddress                                // Wi-Fi BSSID
	SensorBytesSent                                    // Bytes Sent
	SensorBytesRecv                                    // Bytes Received
	SensorBytesSentRate                                // Bytes Sent Throughput
	SensorBytesRecvRate                                // Bytes Received Throughput
	SensorPowerProfile                                 // Power Profile
	SensorBoottime                                     // Last Reboot
	SensorUptime                                       // Uptime
	SensorLoad1                                        // CPU load average (1 min)
	SensorLoad5                                        // CPU load average (5 min)
	SensorLoad15                                       // CPU load average (15 min)
	SensorScreenLock                                   // Screen Lock
	SensorProblem                                      // Problems
	SensorKernel                                       // Kernel Version
	SensorDistribution                                 // Distribution Name
	SensorVersion                                      // Distribution Version
	SensorUsers                                        // Current Users
	SensorDeviceTemp                                   // Temperature
	SensorPowerState                                   // Power State
)

// SensorTypeValue represents the unique type of sensor data being reported. Every
// sensor will have a different type. A SensorTypeValue maps to an entity in Home
// Assistant.
type SensorTypeValue int
