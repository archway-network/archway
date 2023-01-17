FROM golang:1.19.5-alpine3.17 AS go-builder
# arch can be either x86_64 for amd64 or aarch64 for arm
ARG arch=x86_64
ARG libwasmvm_version=v1.1.1
ARG libwasmvm_aarch64_sha=9ecb037336bd56076573dc18c26631a9d2099a7f2b40dc04b6cae31ffb4c8f9a
ARG libwasmvm_amd64_sha=6e4de7ba9bad4ae9679c7f9ecf7e283dd0160e71567c6a7be6ae47c81ebe7f3

# this comes from standard alpine nightly file
#  https://github.com/rust-lang/docker-rust-nightly/blob/master/alpine3.12/Dockerfile
# with some changes to support our toolchain, etc
RUN set -eux; apk add --no-cache ca-certificates build-base;

RUN apk add git
# NOTE: add these to run with LEDGER_ENABLED=true
# RUN apk add libusb-dev linux-headers

WORKDIR /code
COPY . /code/

# See https://github.com/CosmWasm/wasmvm/releases
ADD https://github.com/CosmWasm/wasmvm/releases/download/$libwasmvm_version/libwasmvm_muslc.aarch64.a /lib/libwasmvm_muslc.aarch64.a
ADD https://github.com/CosmWasm/wasmvm/releases/download/$libwasmvm_version/libwasmvm_muslc.x86_64.a /lib/libwasmvm_muslc.x86_64.a
RUN sha256sum /lib/libwasmvm_muslc.aarch64.a | grep $libwasmvm_aarch64_sha
RUN sha256sum /lib/libwasmvm_muslc.x86_64.a | grep $libwasmvm_amd64_sha

# Copy the library you want to the final location that will be found by the linker flag `-lwasmvm_muslc`
RUN cp /lib/libwasmvm_muslc.${arch}.a /lib/libwasmvm_muslc.a

# force it to use static lib (from above) not standard libgo_cosmwasm.so file
RUN LEDGER_ENABLED=false BUILD_TAGS=muslc LINK_STATICALLY=true make build
RUN echo "Ensuring binary is statically linked ..." \
  && (file /code/build/archwayd | grep "statically linked")

# --------------------------------------------------------
FROM alpine:3.17

COPY --from=go-builder /code/build/archwayd /usr/bin/archwayd

WORKDIR /root/.archway

# safety check to ensure deps are correct
RUN archwayd ensure-binary

# rest server
EXPOSE 1317
# tendermint p2p
EXPOSE 26656
# tendermint rpc
EXPOSE 26657

ENTRYPOINT [ "/usr/bin/archwayd" ]

CMD [ "help" ]
