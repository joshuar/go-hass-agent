# Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
# 
# This software is released under the MIT License.
# https://opensource.org/licenses/MIT
FROM fedora:latest

WORKDIR /usr/src/go-hass-agent

RUN sudo dnf -y --setopt=tsflags=nodocs install golang gcc \
    libXcursor-devel libXrandr-devel mesa-libGL-devel \
    libXi-devel libXinerama-devel libXxf86vm-devel && \
    dnf clean all

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go install github.com/matryer/moq@latest
RUN go install golang.org/x/tools/cmd/stringer@latest
RUN go install golang.org/x/text/cmd/gotext@latest
RUN go generate ./...
RUN go build -v -o /usr/bin/go-hass-agent

CMD ["go-hass-agent", "--terminal"]