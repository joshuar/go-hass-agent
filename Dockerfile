# Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
# SPDX-License-Identifier: MIT

FROM docker.io/alpine:3.22.1 AS builder

ARG APPVERSION

WORKDIR /usr/src/app

# Copy go from official image.
COPY --from=golang:1.25.1-alpine /usr/local/go/ /usr/local/go/
ENV PATH="/root/go/bin:/usr/local/go/bin:${PATH}"

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# install build deps
RUN apk update && apk add --no-cache bash libcap git curl libstdc++ libgcc upx

# install bun
RUN curl -fsSL https://bun.com/install | bash

# install and build frontend with bin
RUN export PATH="${HOME}/.bun/bin:${PATH}" && \
    bun install && \
    bun x esbuild ./web/assets/scripts.js --bundle --minify --outdir=./web/content/ && \
    bun x tailwindcss -i ./web/assets/styles.css -o ./web/content/styles.css --minify

# build the binary
ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w -X github.com/joshuar/go-hass-agent/config.AppVersion=$APPVERSION" -o dist/go-hass-agent

# compress binary with upx
RUN upx --best --lzma dist/go-hass-agent

FROM docker.io/alpine:3.22.1

# Add image labels.
LABEL org.opencontainers.image.title="Go Hass Agent"
LABEL org.opencontainers.image.source=https://github.com/joshuar/go-hass-agent
LABEL org.opencontainers.image.description=" A Home Assistant, native app for desktop/laptop devices"
LABEL org.opencontainers.image.licenses=MIT

# Add bash and dbus
RUN apk update && apk add bash dbus dbus-x11 strace

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
CMD ["--terminal", "run"]


