# Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
# SPDX-License-Identifier: 	AGPL-3.0-or-later

# https://hub.docker.com/r/golangci/golangci-lint
FROM docker.io/golangci/golangci-lint:v2.15.0-alpine@sha256:fb5d87d8e469f583363e2547f04e0b0d87f534dfa4617634730b0bd06e53aac5 AS golangci-lint
# https://hub.docker.com/_/golang
FROM docker.io/golang:1.27.3-alpine3.24@sha256:cd410a5c6686417875bc4815688e96ae9769f8562debc545d29fe58e73d174f0

COPY --from=golangci-lint /usr/bin/golangci-lint /usr/local/bin/golangci-lint-v2

# https://github.com/JanDeDobbeleer/oh-my-posh/releases
ARG OMP_VERSION=v31.9.1
# Used for selecting package binaries for install.
ARG TARGETARCH
# Used for installing additional packages.
ARG EXTRA_PKGS
# Add non-root user
ARG USER_NAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID

# Install additional packages
RUN --mount=type=cache,target=/var/cache/apk \
    ln -vs /var/cache/apk /etc/apk/cache \
    && apk add sudo libstdc++ build-base openssh-client ca-certificates tzdata make jq ripgrep less procps curl git \
    bash fish micro dotenv unzip graphviz npm nodejs bash-completion python3 py3-crcmod mosquitto-clients dbus dbus-x11 \
    ${EXTRA_PKGS}

# Install oh-my-posh
RUN curl -fsSL -o /usr/local/bin/oh-my-posh \
    "https://github.com/JanDeDobbeleer/oh-my-posh/releases/download/${OMP_VERSION}/posh-linux-${TARGETARCH}" \
    && chmod +x /usr/local/bin/oh-my-posh


RUN addgroup --gid $USER_GID $USER_NAME \
    && adduser --uid $USER_UID --ingroup $USER_NAME --shell /usr/bin/fish $USER_NAME \
    --disabled-password --gecos "" \
    && mkdir -p /etc/sudoers.d \
    && echo "$USER_NAME ALL=(ALL:ALL) NOPASSWD: ALL" > /etc/sudoers.d/$USER_NAME \
    && chmod 0440 /etc/sudoers.d/$USER_NAME \
    && chown -R $USER_UID:$USER_GID /home/$USER_NAME

ENV XDG_RUNTIME_DIR=/run/user/$USER_UID
RUN mkdir -p $XDG_RUNTIME_DIR && chmod 0700 $XDG_RUNTIME_DIR && chown $USER_UID:$USER_GID $XDG_RUNTIME_DIR
RUN mkdir -p /datahome/fish /datahome/bash && chown -R $USER_NAME /datahome

USER $USER_NAME

ENV _CONTAINERS_USERNS_CONFIGURED=""

# The base container sets XDG_CACHE_HOME XDG_CONFIG_HOME specifically for the root user, we can't unset them in a way
# that vscode will pick up, so we set them to values for the new user. Installing go extensions via vscode use these
# paths so if we just leave it set to /root/.cache we'll get permission errors.
ENV XDG_CONFIG_HOME=/home/$USER_NAME/.config
ENV XDG_CACHE_HOME=/home/$USER_NAME/.cache

# Set up path and environment.
ENV NPM_CONFIG_PREFIX=/home/${USER_NAME}/.npm-global \
    NPM_CONFIG_CACHE=/home/${USER_NAME}/.cache/npm \
    CLOUDSDK_COMPONENT_MANAGER_DISABLE_UPDATE_CHECK=1 \
    CLOUDSDK_CORE_CUSTOM_CA_CERTS_FILE=/etc/ssl/certs/ca-certificates.crt \
    PATH=/usr/local/google-cloud-sdk/bin:/home/${USER_NAME}/.local/bin:/home/${USER_NAME}/.npm-global/bin:$PATH

# Create user-owned dirs up front so cache mounts and volumes inherit correct ownership.
RUN mkdir -p ~/.config/fish ~/.cache/go-build $NPM_CONFIG_CACHE $NPM_CONFIG_PREFIX/bin /go/pkg/mod

# Set up Go.
RUN --mount=type=cache,target=/home/vscode/.cache/go-build,uid=1000,gid=1000 \
    go install golang.org/x/tools/gopls@latest \
    && go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install honnef.co/go/tools/cmd/staticcheck@latest

# Configure npm.
RUN mkdir -p $NPM_CONFIG_PREFIX/bin $NPM_CONFIG_CACHE

# Configure shells
RUN sudo chown -R $USER_UID:$USER_GID /home/$USER_NAME \
    && mkdir -p ~/.config/fish \
    && echo 'source <(npm completion)' >> ~/.bashrc \
    && echo 'eval "$(oh-my-posh init bash)"' >> ~/.bashrc \
    && echo 'oh-my-posh init fish | source' >> ~/.config/fish/config.fish

# Save shell history.
RUN SNIPPET="export PROMPT_COMMAND='history -a' && export HISTFILE=/datahome/bash/.bash_history" \
    && mkdir -p /home/${USER_NAME}/.config/fish \
    && touch /datahome/bash/.bash_history \
    && echo 'set -gx XDG_DATA_HOME /datahome' >> /home/${USER_NAME}/.config/fish/config.fish \
    && echo "$SNIPPET" >> "/home/$USER_NAME/.bashrc"

ENV BASE_CONTAINER=alpine

RUN ls -ld /home/vscode ~ && id

CMD ["fish"]
