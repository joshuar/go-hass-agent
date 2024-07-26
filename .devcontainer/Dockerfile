FROM mcr.microsoft.com/devcontainers/base:ubuntu@sha256:08a925b41755bcff957285d31e9bf5f0b7fc74962f455a929c15dc341b984b6f
RUN <<EOF
apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends mosquitto-clients micro fish ripgrep fzf pre-commit \
    && apt-get -y autoremove && apt-get -y clean && rm -rf /var/lib/apt/lists/*
EOF
