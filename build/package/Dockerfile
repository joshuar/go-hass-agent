# Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
# 
# This software is released under the MIT License.
# https://opensource.org/licenses/MIT
FROM fedora:latest

ARG version

VOLUME [ "/host", "/home/go-hass-agent" ]

ENV HOST_SYS=/host/sys
ENV HOST_PROC=/host/proc

RUN dnf -y install https://github.com/joshuar/go-hass-agent/releases/download/v$version/go-hass-agent-$version-1.x86_64.rpm
RUN dnf -y install dbus

USER go-hass-agent
WORKDIR /home/go-hass-agent

ENTRYPOINT [ "/usr/bin/go-hass-agent", "--terminal" ]
