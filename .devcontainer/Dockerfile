# Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
# SPDX-License-Identifier: MIT

ARG ALPINE_VERSION=3.24.1@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b
ARG GO_VERSION=1.26.5-alpine3.24@sha256:111d79159b2326f7e80c4a4706e1ba166acb0e2611df853955f3621828cd49e8

FROM docker.io/golang:${GO_VERSION} AS golang
FROM docker.io/alpine:${ALPINE_VERSION} AS builder

# Copy go from official image
COPY --from=golang /usr/local/go/ /usr/local/go/

# Install additional packages
RUN apk update && apk add sudo openssh curl git bash fish micro graphviz nodejs npm mosquitto-clients dbus dbus-x11 starship

# Add non-root user
ARG USER_NAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID

RUN addgroup --gid $USER_GID $USER_NAME \
    && adduser --uid $USER_UID --ingroup $USER_NAME --shell /usr/bin/fish $USER_NAME \
    --disabled-password --gecos "" \
    && mkdir -p /etc/sudoers.d \
    && echo "$USER_NAME ALL=(ALL:ALL) NOPASSWD: ALL" > /etc/sudoers.d/$USER_NAME \
    && chmod 0440 /etc/sudoers.d/$USER_NAME

ENV XDG_RUNTIME_DIR=/run/user/$USER_UID
RUN mkdir -p $XDG_RUNTIME_DIR && chmod 0700 $XDG_RUNTIME_DIR && chown $USER_UID:$USER_GID $XDG_RUNTIME_DIR

USER $USER_NAME

ENV _CONTAINERS_USERNS_CONFIGURED=""

# The base container sets XDG_CACHE_HOME XDG_CONFIG_HOME specifically for the root user, we can't unset them in a way
# that vscode will pick up, so we set them to values for the new user. Installing go extensions via vscode use these
# paths so if we just leave it set to /root/.cache we'll get permission errors.
ENV XDG_CONFIG_HOME=/home/$USER_NAME/.config
ENV XDG_CACHE_HOME=/home/$USER_NAME/.cache
# Set up XDG_RUNTIME_DIR
RUN export XDG_RUNTIME_DIR=/tmp/$USER_UID-runtime-dir \
    && mkdir $XDG_RUNTIME_DIR \
    && chmod 0700 $XDG_RUNTIME_DIR
ENV XDG_RUNTIME_DIR=/tmp/$USER_UID-runtime-dir

ENTRYPOINT ["/usr/bin/fish"]
