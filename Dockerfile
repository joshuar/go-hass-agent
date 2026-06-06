# Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
# SPDX-License-Identifier: MIT

# Alpine base.
#
# https://hub.docker.com/_/alpine
ARG ALPINE_VERSION=3.23.4
FROM --platform=$BUILDPLATFORM docker.io/alpine:${ALPINE_VERSION} AS builder

ARG TARGETOS
ARG TARGETARCH
ARG APPVERSION

WORKDIR /usr/src/app

# Copy go from official image.
#
# https://hub.docker.com/_/golang
ARG GO_VERSION=1.26.4-alpine3.23
COPY --from=docker.io/golang:${GO_VERSION} /usr/local/go/ /usr/local/go/
ENV PATH="/root/go/bin:/usr/local/go/bin:/usr/local/bin:${PATH}"

# install build deps
RUN <<EOF
apk add npm upx ca-certificates
EOF

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# install and build frontend with npm (we don't use bun as it is unsupported on some arches we support)
RUN <<EOF
npm clean-install
npm run build:js
npm run build:css
EOF

# build the binary
ENV CGO_ENABLED=0
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w -X github.com/joshuar/go-hass-agent/config.AppVersion=$APPVERSION" -o dist/go-hass-agent

# compress binary with upx
RUN upx --best --lzma dist/go-hass-agent

FROM docker.io/alpine:${ALPINE_VERSION}

# Don't log to a file when running in a container
ENV GOHASSAGENT_NOLOGFILE=1

# Add image labels.
LABEL org.opencontainers.image.title="Go Hass Agent"
LABEL org.opencontainers.image.source=https://github.com/joshuar/go-hass-agent
LABEL org.opencontainers.image.description="A Home Assistant, native app for desktop/laptop devices"
LABEL org.opencontainers.image.licenses=MIT

# Add bash and dbus
RUN apk update && apk add bash dbus dbus-x11

# Copy binary over from builder stage
COPY --from=builder /usr/src/app/dist/go-hass-agent /usr/bin/go-hass-agent

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
CMD ["run"]


