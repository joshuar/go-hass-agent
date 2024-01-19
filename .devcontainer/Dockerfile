# Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
# 
# This software is released under the MIT License.
# https://opensource.org/licenses/MIT

FROM mcr.microsoft.com/devcontainers/base:ubuntu
# Install the xz-utils package
RUN sudo apt -y update  && export DEBIAN_FRONTEND=noninteractive && \
    sudo apt -y install gcc libgl1-mesa-dev xorg-dev dbus-x11 desktop-file-utils fish micro && \
    rm -rf /var/lib/apt/lists/*
