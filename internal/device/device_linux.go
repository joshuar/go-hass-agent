// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package device

import (
	"encoding/json"
	"os/exec"
	"os/user"
	"strings"

	"git.lukeshu.com/go/libsystemd/sd_id128"
	"github.com/acobaugh/osrelease"
	"github.com/rs/zerolog/log"
)

const (
	Name    = "go-hass-agent"
	Version = "0.0.1"
)

type linuxDevice struct {
	hostnameCtl *hostnameCtl
	osRelease   map[string]string
	appID       string
	machineID   string
}

type hostnameCtl struct {
	Hostname                  string `json:"Hostname"`
	StaticHostname            string `json:"StaticHostname"`
	PrettyHostname            string `json:"PrettyHostname"`
	DefaultHostname           string `json:"DefaultHostname"`
	HostnameSource            string `json:"HostnameSource"`
	IconName                  string `json:"IconName"`
	Chassis                   string `json:"Chassis"`
	Deployment                string `json:"Deployment"`
	Location                  string `json:"Location"`
	KernelName                string `json:"KernelName"`
	KernelRelease             string `json:"KernelRelease"`
	KernelVersion             string `json:"KernelVersion"`
	OperatingSystemPrettyName string `json:"OperatingSystemPrettyName"`
	OperatingSystemCPEName    string `json:"OperatingSystemCPEName"`
	OperatingSystemHomeURL    string `json:"OperatingSystemHomeURL"`
	HardwareVendor            string `json:"HardwareVendor"`
	HardwareModel             string `json:"HardwareModel"`
	HardwareSerial            string `json:"HardwareSerial"`
	FirmwareVersion           string `json:"FirmwareVersion"`
	ProductUUID               string `json:"ProductUUID"`
}

func (l *linuxDevice) AppName() string {
	return Name
}

func (l *linuxDevice) AppVersion() string {
	return Version
}

func (l *linuxDevice) AppID() string {
	return l.appID
}

func (l *linuxDevice) DeviceName() string {
	shortHostname, _, _ := strings.Cut(l.hostnameCtl.Hostname, ".")
	return shortHostname
}

func (l *linuxDevice) DeviceID() string {
	return l.machineID
}

func (l *linuxDevice) Manufacturer() string {
	return l.hostnameCtl.HardwareVendor
}

func (l *linuxDevice) Model() string {
	return l.hostnameCtl.HardwareModel
}

func (l *linuxDevice) OsName() string {
	return l.hostnameCtl.OperatingSystemPrettyName
}

func (l *linuxDevice) OsVersion() string {
	return l.osRelease["VERSION_ID"]
}

func (l *linuxDevice) SupportsEncryption() bool {
	return false
}

func (l *linuxDevice) AppData() interface{} {
	return &struct {
		PushWebsocket bool `json:"push_websocket_channel"`
	}{
		PushWebsocket: true,
	}
}

func NewDevice() *linuxDevice {

	hostnameCtlCmd := checkForBinary("hostnamectl")
	out, err := exec.Command(hostnameCtlCmd, "--json=short").Output()
	if err != nil {
		log.Fatal().Caller().
			Msgf("Could not execute hostnamectl: %v", err)
	}
	var h *hostnameCtl
	err = json.Unmarshal(out, &h)
	if err != nil {
		log.Fatal().Caller().
			Msgf("Failed to parse output of hostnamectl: %v", err)
	}

	osrelease, err := osrelease.Read()
	if err != nil {
		log.Fatal().Caller().
			Msgf("Unable to read file /etc/os-release: %v", err)
	}

	currentUser, err := user.Current()
	if err != nil {
		log.Fatal().Caller().
			Msgf("Could not retrieve current user details: %v", err.Error())
	}

	machineID, err := sd_id128.GetRandomUUID()
	if err != nil {
		log.Fatal().Caller().
			Msgf("Could not retrieve a machine ID: %v", err)
	}

	return &linuxDevice{
		hostnameCtl: h,
		osRelease:   osrelease,
		appID:       Name + "-" + currentUser.Username,
		machineID:   machineID.String(),
	}
}

func checkForBinary(binary string) string {
	path, err := exec.LookPath(binary)
	if err != nil {
		log.Fatal().Caller().
			Msgf("Could not find needed executable %s in PATH", binary)
	}
	return path
}
