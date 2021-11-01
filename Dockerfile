# docker build . -t tendermint-mpc/validator:latest
# docker run --rm -it tendermint-mpc/validator:latest /bin/sh

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

COPY --from=builder /code/build/key2shares /usr/bin/key2shares
COPY --from=builder /code/build/signer /usr/bin/signer


CMD ["echo", "image build successfully"]



