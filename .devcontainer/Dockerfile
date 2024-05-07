FROM mcr.microsoft.com/devcontainers/base:ubuntu
RUN <<EOF
sudo dpkg --add-architecture armhf
sudo dpkg --add-architecture arm64
sudo sed -i 's|^deb\s|deb [arch=amd64] |g' /etc/apt/sources.list
cat /etc/apt/sources.list
EOF
COPY <<EOF /etc/apt/sources.list.d/ports.list
deb [arch=armhf,arm64] http://ports.ubuntu.com/ubuntu-ports/ jammy main
#deb-src [arch=armhf,arm64] http://ports.ubuntu.com/ubuntu-ports/ jammy main
deb [arch=armhf,arm64] http://ports.ubuntu.com/ubuntu-ports/ jammy-updates main
#deb-src [arch=armhf,arm64] http://ports.ubuntu.com/ubuntu-ports/ jammy-updates main
deb [arch=armhf,arm64] http://ports.ubuntu.com/ubuntu-ports/ jammy-security main
#deb-src [arch=armhf,arm64] http://ports.ubuntu.com/ubuntu-ports/ jammy-security main
EOF
RUN sudo apt-get -y update
ARG DEBIAN_FRONTEND noninteractive
RUN <<EOF
sudo apt-get -y install desktop-file-utils micro fish mosquitto-clients gcc gcc-arm-linux-gnueabihf gcc-aarch64-linux-gnu
sudo apt-get -y install libgl1-mesa-dev:armhf xserver-xorg-dev:armhf libxcursor-dev:armhf libxrandr-dev:armhf libxinerama-dev:armhf libxi-dev:armhf libgl1-mesa-dev:armhf libxxf86vm-dev:armhf  
sudo apt-get -y install libgl1-mesa-dev:arm64 xserver-xorg-dev:arm64 libxcursor-dev:arm64 libxrandr-dev:arm64 libxinerama-dev:arm64 libxi-dev:arm64 libgl1-mesa-dev:arm64 libxxf86vm-dev:arm64
sudo apt-get -y install libgl1-mesa-dev:amd64 xserver-xorg-dev:amd64 libxcursor-dev:amd64 libxrandr-dev:amd64 libxinerama-dev:amd64 libxi-dev:amd64 libgl1-mesa-dev:amd64 libxxf86vm-dev:amd64
sudo rm -rf /var/lib/apt/lists/* /var/cache/apt/*
EOF
