# Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
# SPDX-License-Identifier: MIT

ARG GO_VERSION

# Alpine base.
# https://hub.docker.com/_/alpine
FROM docker.io/alpine:3.23.0@sha256:51183f2cfa6320055da30872f211093f9ff1d3cf06f39a0bdb212314c5dc7375 AS builder

# Copy go from official image
COPY --from=docker.io/golang:1.25.6-alpine@sha256:98e6cffc31ccc44c7c15d83df1d69891efee8115a5bb7ede2bf30a38af3e3c92 /usr/local/go/ /usr/local/go/

# Install additional packages
RUN apk update && apk add sudo openssh curl git bash fish micro graphviz nodejs npm mosquitto-clients

# Install starship
RUN cd /tmp && curl -sS https://starship.rs/install.sh | sh -s -- -y || exit -1

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

USER $USER_NAME

RUN <<EOF
    mkdir -p ~/.config/fish
    echo 'set --export PATH "/workspace/node_modules/.bin $PATH' >> ~/.config/fish/config.fish
    echo 'set --export PATH "$HOME/go/bin" /go/bin /usr/local/go/bin $PATH' >> ~/.config/fish/config.fish
EOF

# The base container sets XDG_CACHE_HOME XDG_CONFIG_HOME specifically for the root user, we can't unset them in a way that vscode will pick up, so we set them to values for the new user.
# Installing go extensions via vscode use these paths so if we just leave it set to /root/.cache we'll get permission errors.
ENV XDG_CONFIG_HOME=/home/$USER_NAME/.config
ENV XDG_CACHE_HOME=/home/$USER_NAME/.cache

ENTRYPOINT ["/usr/bin/fish"]
