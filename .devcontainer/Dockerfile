# Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
# SPDX-License-Identifier: MIT

ARG GO_VERSION

# Alpine base.
# https://hub.docker.com/_/alpine
FROM docker.io/alpine:3.23.3@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659 AS builder

# Copy go from official image
COPY --from=docker.io/golang:1.26.1-alpine-3.23@sha256:2389ebfa5b7f43eeafbd6be0c3700cc46690ef842ad962f6c5bd6be49ed82039 /usr/local/go/ /usr/local/go/

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

# The base container sets XDG_CACHE_HOME XDG_CONFIG_HOME specifically for the root user, we can't unset them in a way that vscode will pick up, so we set them to values for the new user.
# Installing go extensions via vscode use these paths so if we just leave it set to /root/.cache we'll get permission errors.
ENV XDG_CONFIG_HOME=/home/$USER_NAME/.config
ENV XDG_CACHE_HOME=/home/$USER_NAME/.cache
RUN mkdir -p $XDG_CONFIG_HOME $XDG_CACHE_HOME

ENTRYPOINT ["/usr/bin/fish"]
