FROM --platform=$BUILDPLATFORM golang:alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers git wget
ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG LINK_STATICALLY=true
ARG COSNWAMS_VERSION=1.2.3
ENV LINK_STATICALLY=${LINK_STATICALLY}

COPY . /usr/src/archway

# get cosmwasm
RUN wget -q https://github.com/CosmWasm/wasmvm/releases/download/v${COSNWAMS_VERSION}/libwasmvm_muslc.aarch64.a -O /usr/lib/libwasmvm.aarch64.a && \
    wget -q https://github.com/CosmWasm/wasmvm/releases/download/v${COSNWAMS_VERSION}/libwasmvm_muslc.x86_64.a -O /usr/lib/libwasmvm.x86_64.a

WORKDIR /usr/src/archway

RUN make build

FROM --platform=$TARGETPLATFORM alpine:latest

# copy archwayd binary
COPY --from=builder /usr/src/archway/build/archwayd /usr/bin/archwayd

# copy tls cert
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /root/.archway

# rest server, tendermint p2p, tendermint rpc
EXPOSE 1317 26656 26657


ENTRYPOINT [ "/usr/bin/archwayd" ]

VOLUME [ "/contracts", "/opt" ]

CMD [ "help" ]
