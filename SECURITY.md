<!--
 Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
 SPDX-License-Identifier: MIT
-->

# Security Policy

Thanks for helping to make Go Hass Agent a safe and useful application for everyone.

## Supported Versions

Only the latest released version of Go Hass Agent will be supported with security updates.

## Reporting a Vulnerability

Security issues and vulnerabilities can be reported privately by following the
GitHub documentation: [Privately reporting a security
vulnerability](https://docs.github.com/en/code-security/security-advisories/guidance-on-reporting-and-writing-information-about-vulnerabilities/privately-reporting-a-security-vulnerability#privately-reporting-a-security-vulnerability).

**Please do not report security vulnerabilities through public GitHub issues,
discussions, or pull requests.**

Please include as much of the information listed below as you can to help us
better understand and resolve the issue:

- The type of issue (e.g., buffer overflow, SQL injection, or cross-site scripting)
- Full paths of source file(s) related to the manifestation of the issue
- The location of the affected source code (tag/branch/commit or direct URL)
- Any special configuration required to reproduce the issue
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue, including how an attacker might exploit the issue

This information will help us triage your report more quickly.

Security issues and vulnerabilities will be addressed with reasonable effort but no guarantees are made with regards to
resolution of reports within any time frame or a fix at all.

## Permissions and Capabilities

### Cannot be run as root user

Go Hass Agent cannot be run as root or a user with effective root permissions. Go Hass Agent will detect this situation
and refuse to start.

### Arbitrary script/commands

Some features provide the ability to execute arbitrary scripts and commands on the device running the agent:

- [custom script sensors](./README.md#other-custom-commands).
- [MQTT commands](./README.md#-mqtt-sensors-and-controls).

These may or may not represent a significant security issue for some users. They are not enabled by default and require
manual configuration to use.

### Sensors requiring additional capabilities

Some sensors require additional capabilities on the Go Hass Agent binary:

- SMART disk monitoring: requires `cap_sys_rawio,cap_sys_admin,cap_mknod,cap_dac_override=+ep`

When installed via packages (RPM/DEB/ARCH) or using the [official container
image](https://github.com/joshuar/go-hass-agent/pkgs/container/go-hass-agent), the binary will have the required
capabilities by default.

If this is not desired, the binary should be modified or a custom image used that removes these capabilities. This will
of course result in the sensors requiring those capabilities to be unavailable, but otherwise Go Hass Agent will
continue to run. For example with `setcap -r /path/to/go-hass-agent`.
