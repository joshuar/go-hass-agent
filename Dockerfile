# Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
# 
# This software is released under the MIT License.
# https://opensource.org/licenses/MIT

ARG GO_VERSION=1.22

FROM docker.io/golang:${GO_VERSION} AS builder
WORKDIR /usr/src/go-hass-agent

# Default to an amd64 build
ARG BUILD_ARCH=amd64

# copy the src to the workdir
ADD . .
# install mage
RUN go install github.com/magefile/mage@latest
# install build dependencies
RUN mage -v -d build/magefiles -w . preps:deps ${BUILD_ARCH}
# build the binary
RUN mage -v -d build/magefiles -w . build:full ${BUILD_ARCH}

FROM ubuntu
# copy binary over from builder stage
COPY --from=builder /usr/src/go-hass-agent/dist/go-hass-agent-* /usr/bin/go-hass-agent
# reinstall minimum libraries for running
RUN mkdir /etc/dpkg/dpkg.conf.d
COPY <<EOF /etc/dpkg/dpkg.conf.d/excludes
# Drop all man pages
path-exclude=/usr/share/man/*
# Drop all translations
path-exclude=/usr/share/locale/*/LC_MESSAGES/*.mo
# Drop all documentation ...
path-exclude=/usr/share/doc/*
# ... except copyright files ...
path-include=/usr/share/doc/*/copyright
# ... and Debian changelogs for native & non-native packages
path-include=/usr/share/doc/*/changelog.*
EOF
RUN <<EOF
export DEBIAN_FRONTEND=noninteractive
apt-get -y update 
apt-get -y install libgl1 libx11-6 libglx0 libglvnd0 libxcb1 libxau6 libxdmcp6 dbus-x11 
rm -rf /var/lib/apt/lists/* /var/cache/apt/*
EOF
# set up run entrypoint/cmd
ENTRYPOINT ["go-hass-agent"]
CMD ["--terminal", "run"]
