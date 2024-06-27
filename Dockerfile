# Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
#
# This software is released under the MIT License.
# https://opensource.org/licenses/MIT

FROM --platform=$BUILDPLATFORM ubuntu@sha256:94db6b944510db19c0ff5eb13281cf166abfe6f9e01a6f8e716e976664537c60 AS builder
# add ca-certificates so go command can download stuff
RUN <<EOF
export DEBIAN_FRONTEND=noninteractive
apt-get -y update
apt-get -y install ca-certificates
EOF
# download and install go
ADD https://go.dev/dl/go1.22.4.linux-amd64.tar.gz /tmp/go1.22.4.linux-amd64.tar.gz
RUN rm -rf /usr/local/go && tar -C /usr/local -xzf /tmp/go1.22.4.linux-amd64.tar.gz && rm /tmp/go1.22.4.linux-amd64.tar.gz
ENV PATH="$PATH:/usr/local/go/bin:/root/go/bin"

WORKDIR /usr/src/go-hass-agent
# import BUILDPLATFORM
ARG BUILDPLATFORM
# copy the src to the workdir
ADD . .
# install mage
RUN go install github.com/magefile/mage@v1.15.0
# install build dependencies
RUN mage -v -d build/magefiles -w . preps:deps
# build the binary
RUN mage -v -d build/magefiles -w . build:full

FROM --platform=$BUILDPLATFORM ubuntu@sha256:94db6b944510db19c0ff5eb13281cf166abfe6f9e01a6f8e716e976664537c60
# import BUILDPLATFORM and TARGETARCH
ARG BUILDPLATFORM
ARG TARGETARCH
# copy binary over from builder stage
COPY --from=builder /usr/src/go-hass-agent/dist/go-hass-agent-$TARGETARCH* /usr/bin/go-hass-agent
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
# install runtime deps
COPY --from=builder /usr/src/go-hass-agent/build/scripts/install-run-deps /tmp/install-run-deps
COPY --from=builder /usr/src/go-hass-agent/build/scripts/enable-multiarch /tmp/enable-multiarch
RUN /tmp/enable-multiarch $BUILDPLATFORM && rm /tmp/enable-multiarch
RUN /tmp/install-run-deps $BUILDPLATFORM && rm /tmp/install-run-deps
# set up run entrypoint/cmd
ENTRYPOINT ["go-hass-agent"]
CMD ["--terminal", "run"]
