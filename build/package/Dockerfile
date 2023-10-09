# Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
# 
# This software is released under the MIT License.
# https://opensource.org/licenses/MIT

ARG server="http://localhost:8123"
ARG token
ARG version

FROM fedora:latest

RUN dnf -y install https://github.com/joshuar/go-hass-agent/releases/download/v$version/go-hass-agent-$version-1.x86_64.rpm

RUN useradd -ms /bin/bash go-hass-agent -G wheel
RUN echo "%wheel ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/999-custom
USER go-hass-agent
WORKDIR /home/go-hass-agent
RUN go-hass-agent --terminal register --server $server --token $token

ENTRYPOINT [ "bash" ]
CMD ["go-hass-agent","--terminal"]
