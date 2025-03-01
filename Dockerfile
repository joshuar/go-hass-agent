# Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
#
# This software is released under the MIT License.
# https://opensource.org/licenses/MIT

FROM docker.io/alpine@sha256:0a4eaa0eecf5f8c050e5bba433f58c052be7587ee8af3e8b3910ef9ab5fbe9f5 AS builder
# Copy go from official image.
COPY --from=golang:1.24.0-alpine /usr/local/go/ /usr/local/go/
ENV PATH="/root/go/bin:/usr/local/go/bin:${PATH}"
# Import TARGETPLATFORM.
ARG TARGETPLATFORM
# set the workdir
WORKDIR /usr/src/go-hass-agent
# copy the src to the workdir
ADD . .
# install bash
RUN apk update && apk add bash
# install build dependencies
RUN go run github.com/magefile/mage -d build/magefiles -w . preps:deps
# build the binary
RUN go run github.com/magefile/mage -d build/magefiles -w . build:full

FROM docker.io/alpine@sha256:0a4eaa0eecf5f8c050e5bba433f58c052be7587ee8af3e8b3910ef9ab5fbe9f5
# Add image labels.
LABEL org.opencontainers.image.source=https://github.com/joshuar/go-hass-agent
LABEL org.opencontainers.image.description=" A Home Assistant, native app for desktop/laptop devices"
LABEL org.opencontainers.image.licenses=MIT
# Import TARGETPLATFORM and TARGETARCH
ARG TARGETPLATFORM
ARG TARGETARCH
# Add bash and dbus
RUN apk update && apk add bash dbus dbus-x11
# Install run deps
COPY --from=builder /usr/src/go-hass-agent/build/scripts/install-run-deps /tmp/install-run-deps
RUN /tmp/install-run-deps $TARGETPLATFORM && rm /tmp/install-run-deps
# Copy binary over from builder stage
COPY --from=builder /usr/src/go-hass-agent/dist/go-hass-agent-$TARGETARCH* /usr/bin/go-hass-agent
# Allow custom uid and gid
ARG UID=1000
ARG GID=1000
# Add user
RUN addgroup --gid "${GID}" go-hass-agent && \
    adduser --disabled-password --gecos "" --ingroup go-hass-agent \
    --uid "${UID}" go-hass-agent
USER go-hass-agent
# Set up run entrypoint/cmd
ENTRYPOINT ["go-hass-agent"]
CMD ["--terminal", "run"]
