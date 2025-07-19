# Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
# SPDX-License-Identifier: MIT

#!/bin/sh

cleanInstall() {
    # Set capabilities required for some workers if `setcap` is available. If not, some workers will not be able to run.
    if type setcap >/dev/null; then
        setcap 'cap_sys_rawio=ep cap_sys_admin=ep cap_mknod=ep' /usr/bin/go-hass-agent
    fi
}

# upgrade() {
# }

# Step 2, check if this is a clean install or an upgrade
action="$1"
if  [ "$1" = "configure" ] && [ -z "$2" ]; then
  # Alpine linux does not pass args, and deb passes $1=configure
  action="install"
elif [ "$1" = "configure" ] && [ -n "$2" ]; then
    # deb passes $1=configure $2=<current version>
    action="upgrade"
fi

case "$action" in
  "1" | "install")
    cleanInstall
    ;;
#   "2" | "upgrade")
#     upgrade
#     ;;
  *)
    # $1 == version being installed
    printf "\033[32m Alpine\033[0m"
    cleanInstall
    ;;
esac

