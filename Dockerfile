# docker build . -t dautt/valink:v2.0.0
# docker run --rm -it dautt/valink:v2.0.0 /bin/sh

FROM golang:1.16-alpine3.12 AS builder

# Set up dependencies
ENV PACKAGES make gcc git libc-dev bash linux-headers eudev-dev jq

WORKDIR /code
COPY . /code/


# Install minimum necessary dependencies
RUN apk add --no-cache $PACKAGES

RUN make build

# ----------------------------

FROM alpine:3.12

COPY --from=builder /code/build/valink /usr/bin/valink

CMD ["echo", "image build successfully"]



