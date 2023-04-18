![](https://github.com/archway-network/archway/blob/main/banner.png)
# Archway

[![Version](https://img.shields.io/github/v/tag/archway-network/archway.svg?sort=semver&style=flat-square)](https://github.com/archway-network/archway/releases/latest)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue?style=flat-square&logo=go)](https://pkg.go.dev/github.com/archway-network/archway)
[![Go Report Card](https://goreportcard.com/badge/github.com/archway-network/archway)](https://goreportcard.com/report/github.com/archway-network/archway)
[![codecov](https://codecov.io/gh/archway-network/archway/branch/master/graph/badge.svg)](https://codecov.io/gh/archway-network/archway)
[![License:Apache-2.0](https://img.shields.io/github/license/archway-network/archway.svg?style=flat-square)](https://github.com/archway-network/archway/LICENSE)


The core implementation of the Archway protocol leverages the [Cosmos SDK](https://cosmos.network) and [CosmWasm](https://cosmwasm.com) to reward validators and creators for their contributions to the network.

## System Requirements

The following specifications have been found to work well:

- An x86-64 (amd64) multi-core CPU (AMD / Intel);
    - Higher clock speeds are preferred as Tendermint is mostly single-threaded;
- 64GB RAM;
- 1TB NVMe SSD Storage (disk i/o is crucial);
- 100Mbps bi-directional Internet connection;

## Software Dependencies

The following software should be installed on the target system:

- The Go Programming Language (https://go.dev)
- Git Distributed Version Control (https://git-scm.com)
- Docker (https://www.docker.com)
- GNU Make (https://www.gnu.org/software/make)

## Build from Source

[Clone the repository](https://github.com/archway-network/archway), checkout the `main` branch and build:

```sh
cd archway
git checkout main
make install
```

This will install the `archwayd` binary to your `GOPATH`.

## Dockerized Container

A docker image is also provided for test setups.

```bash
docker build -t archwaynetwork/archwayd:latest .
```

_Tip: Make sure to include the __dot__ from the code above ^^^_

## Documentation

To learn more, please [visit the official Archway documentation](https://docs.archway.io).
