// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

//go:generate stringer -type=sensorType -output sensorTypeStrings.go -linecomment
const (
	appActive         sensorType = iota + 1 // Active App
	appRunning                              // Running Apps
	battType                                // Battery Type
	battPercentage                          // Battery Level
	battTemp                                // Battery Temperature
	battVoltage                             // Battery Voltage
	battEnergy                              // Battery Energy
	battEnergyRate                          // Battery Power
	battState                               // Battery State
	battNativePath                          // Battery Path
	battLevel                               // Battery Level
	battModel                               // Battery Model
	memTotal                                // Memory Total
	memAvail                                // Memory Available
	memUsed                                 // Memory Used
	swapTotal                               // Swap Memory Total
	swapUsed                                // Swap Memory Used
	swapFree                                // Swap Memory Free
	connectionState                         // Connection State
	connectionID                            // Connection ID
	connectionDevices                       // Connection Device
	connectionType                          // Connection Type
	connectionIPv4                          // Connection IPv4
	connectionIPv6                          // Connection IPv6
	addressIPv4                             // IPv4 Address
	addressIPv6                             // IPv6 Address
	wifiSSID                                // Wi-Fi SSID
	wifiFrequency                           // Wi-Fi Frequency
	wifiSpeed                               // Wi-Fi Link Speed
	wifiStrength                            // Wi-Fi Signal Strength
	wifiHWAddress                           // Wi-Fi BSSID
	bytesSent                               // Bytes Sent
	bytesRecv                               // Bytes Received
	bytesTx                                 // Upload Throughput
	bytesRx                                 // Download Throughput
	powerProfile                            // Power Profile
	boottime                                // Last Reboot
	uptime                                  // Uptime
	load1                                   // CPU load average (1 min)
	load5                                   // CPU load average (5 min)
	load15                                  // CPU load average (15 min)
	screenLock                              // Screen Lock
	problem                                 // Problems
	kernel                                  // Kernel Version
	distribution                            // Distribution Name
	version                                 // Distribution Version
	users                                   // Current Users
	deviceTemp                              // Temperature
)

// sensorType represents the unique type of sensor data being reported. Every
// sensor will have a different type. A sensorType maps to an entity in Home
// Assistant.
type sensorType int
