FROM mcr.microsoft.com/devcontainers/base:ubuntu
# Install the xz-utils package
RUN sudo apt-get -y update  && export DEBIAN_FRONTEND=noninteractive && \
    sudo apt-get -y install gcc libgl1-mesa-dev xorg-dev dbus-x11 desktop-file-utils fish micro mosquitto-clients && \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*
