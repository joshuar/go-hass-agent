# Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
#
# This software is released under the MIT License.
# https://opensource.org/licenses/MIT

FROM docker.io/alpine@sha256:0a4eaa0eecf5f8c050e5bba433f58c052be7587ee8af3e8b3910ef9ab5fbe9f5 AS builder
# import TARGETPLATFORM
ARG TARGETPLATFORM
# set the workdir
WORKDIR /usr/src/go-hass-agent
# copy the src to the workdir
ADD . .
ENV PATH="$PATH:/root/go/bin"
# install go, bash
RUN apk update && apk add go bash
# install mage
RUN go install github.com/magefile/mage@v1.15.0
# install build dependencies
RUN mage -v -d build/magefiles -w . preps:deps
# build the binary
RUN mage -v -d build/magefiles -w . build:full

FROM docker.io/alpine@sha256:0a4eaa0eecf5f8c050e5bba433f58c052be7587ee8af3e8b3910ef9ab5fbe9f5
# import TARGETPLATFORM and TARGETARCH
ARG TARGETPLATFORM
ARG TARGETARCH
# add bash
RUN apk update && apk add bash
# install run deps
COPY --from=builder /usr/src/go-hass-agent/build/scripts/install-run-deps /tmp/install-run-deps
RUN /tmp/install-run-deps $TARGETPLATFORM && rm /tmp/install-run-deps
# copy binary over from builder stage
COPY --from=builder /usr/src/go-hass-agent/dist/go-hass-agent-$TARGETARCH* /usr/bin/go-hass-agent
# set up run entrypoint/cmd
ENTRYPOINT ["go-hass-agent"]
CMD ["--terminal", "run"]
