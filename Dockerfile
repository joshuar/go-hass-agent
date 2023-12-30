# Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
# 
# This software is released under the MIT License.
# https://opensource.org/licenses/MIT
FROM golang:1.21

WORKDIR /usr/src/go-hass-agent

ENV DEBIAN_FRONTEND=noninteractive
RUN apt update && apt -y install golang gcc libgl1-mesa-dev xorg-dev && rm -rf /var/lib/apt/lists/*

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go install github.com/matryer/moq@latest
RUN go install golang.org/x/tools/cmd/stringer@latest
RUN go install golang.org/x/text/cmd/gotext@latest
RUN go generate ./...
RUN go build -v -o /go/bin/go-hass-agent

CMD ["go-hass-agent", "--terminal"]