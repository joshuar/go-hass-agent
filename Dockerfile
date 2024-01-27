# Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
# 
# This software is released under the MIT License.
# https://opensource.org/licenses/MIT
FROM golang:1.21

WORKDIR /usr/src/go-hass-agent

# https://developer.fyne.io/started/#prerequisites
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get -y install gcc libgl1-mesa-dev xorg-dev && rm -rf /var/lib/apt/lists/*

# copy the src to the workdir
ADD . .

# install build dependencies
RUN go install github.com/matryer/moq@latest && \
  go install golang.org/x/tools/cmd/stringer@latest && \
  go install golang.org/x/text/cmd/gotext@latest

# create the VERSION file
WORKDIR /usr/src/go-hass-agent/internal/agent/config
RUN printf %s $(git tag | tail -1) > VERSION

WORKDIR /usr/src/go-hass-agent

# build the binary
RUN go generate ./... && \
  go build -v -o /go/bin/go-hass-agent && \
  rm -fr /usr/src/go-hass-agent

# create a user to run the agent
RUN useradd -ms /bin/bash gouser
USER gouser
WORKDIR /home/gouser

ENTRYPOINT ["go-hass-agent"]
CMD ["--terminal"]
