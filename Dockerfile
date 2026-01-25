# Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
# SPDX-License-Identifier: MIT

# Alpine base.
#
# https://hub.docker.com/_/alpine
FROM --platform=$BUILDPLATFORM docker.io/alpine:3.23.2@sha256:865b95f46d98cf867a156fe4a135ad3fe50d2056aa3f25ed31662dff6da4eb62 AS builder

ARG TARGETOS
ARG TARGETARCH
ARG APPVERSION

WORKDIR /usr/src/app

# Copy go from official image.
#
# https://hub.docker.com/_/golang
COPY --from=docker.io/golang:1.25.5-alpine@sha256:26111811bc967321e7b6f852e914d14bede324cd1accb7f81811929a6a57fea9 /usr/local/go/ /usr/local/go/
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
npm install
npm run build:js
npm run build:css
EOF

# build the binary
ENV CGO_ENABLED=0
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w -X github.com/joshuar/go-hass-agent/config.AppVersion=$APPVERSION" -o dist/go-hass-agent

# compress binary with upx
RUN upx --best --lzma dist/go-hass-agent

FROM docker.io/alpine:3.23.2@sha256:865b95f46d98cf867a156fe4a135ad3fe50d2056aa3f25ed31662dff6da4eb62

# Don't log to a file when running in a container
ENV GOHASSAGENT_NOLOGFILE=1

# Add image labels.
LABEL org.opencontainers.image.title="Go Hass Agent"
LABEL org.opencontainers.image.source=https://github.com/joshuar/go-hass-agent
LABEL org.opencontainers.image.description=" A Home Assistant, native app for desktop/laptop devices"
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


