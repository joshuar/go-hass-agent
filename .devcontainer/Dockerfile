FROM mcr.microsoft.com/devcontainers/base:ubuntu
RUN ../build/scripts/enable-multiarch all
RUN ../build/scripts/install-deps arm arm64 amd64
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install fish micro pre-commit
