FROM mcr.microsoft.com/devcontainers/base:ubuntu
RUN sudo apt-get -y update  && export DEBIAN_FRONTEND=noninteractive && \
    sudo apt-get -y install desktop-file-utils gcc \
        libgl1-mesa-dev xserver-xorg-dev \
        libxcursor-dev libxrandr-dev \
        libxinerama-dev libxi-dev \
        libgl1-mesa-dev libxxf86vm-dev && \
        rm -rf /var/lib/apt/lists/* /var/cache/apt/*
