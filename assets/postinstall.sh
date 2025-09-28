# Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
# SPDX-License-Identifier: MIT

#!/bin/sh

# Set capabilities required for some workers if `setcap` is available. If not, some workers will not be able to run.
setCapabilities() {
    if type setcap >/dev/null; then
        setcap 'cap_sys_rawio,cap_sys_admin,cap_mknod,cap_dac_override=+ep' /usr/bin/go-hass-agent
    fi
}

setCapabilities
