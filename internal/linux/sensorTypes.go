// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

//go:generate stringer -type=sensorType -output sensorTypeStrings.go -linecomment
const (
	appActive         sensorType = iota // Active App
	appRunning                          // Running Apps
	battType                            // Battery Type
	battPercentage                      // Battery Level
	battTemp                            // Battery Temperature
	battVoltage                         // Battery Voltage
	battEnergy                          // Battery Energy
	battEnergyRate                      // Battery Power
	battState                           // Battery State
	battNativePath                      // Battery Path
	battLevel                           // Battery Level
	battModel                           // Battery Model
	memTotal                            // Memory Total
	memAvail                            // Memory Available
	memUsed                             // Memory Used
	swapTotal                           // Swap Memory Total
	swapUsed                            // Swap Memory Used
	swapFree                            // Swap Memory Free
	connectionState                     // Connection State
	connectionID                        // Connection ID
	connectionDevices                   // Connection Device
	connectionType                      // Connection Type
	connectionIPv4                      // Connection IPv4
	connectionIPv6                      // Connection IPv6
	addressIPv4                         // IPv4 Address
	addressIPv6                         // IPv6 Address
	wifiSSID                            // Wi-Fi SSID
	wifiFrequency                       // Wi-Fi Frequency
	wifiSpeed                           // Wi-Fi Link Speed
	wifiStrength                        // Wi-Fi Signal Strength
	wifiHWAddress                       // Wi-Fi BSSID
	bytesSent                           // Bytes Sent
	bytesRecv                           // Bytes Recieved
	powerProfile                        // Power Profile
	boottime                            // Last Reboot
	uptime                              // Uptime
)

// sensorType represents the unique type of sensor data being reported. Every
// sensor will have a different type. A sensorType maps to an entity in Home
// Assistant.
type sensorType int
