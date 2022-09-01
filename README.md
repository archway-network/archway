![](https://github.com/archway-network/archway/blob/main/banner.png)
# Archway

The core implementation of the Archway protocol leverages the [Cosmos SDK](https://cosmos.network) and [CosmWasm](https://cosmwasm.com) to reward validators and creators for their contributions to the network.

# Installation

## Install Golang

Go 1.18 is required for Archway.

If you haven't already, download and install Go. See the official [go.dev documentation](https://golang.org/doc/install). Make sure your `GOBIN` and `GOPATH` are setup.

## Get the Archway source code

Retrieve the source code from the official [archway-network/archway](https://github.com/archway-network/archway) GitHub repository.

```
git clone https://github.com/archway-network/archway
cd archway
git checkout main
```

## Build the Archway binary

You can build with:

```
make install
```

This command installs the `archwayd` to your `GOPATH`.

## Dockerized

A docker image is provided to help with test setups. 

To build a docker image:

```
docker build -t archwaynetwork/archwayd:latest .
```

## Documentation

To learn more, see the official [Archway documentation](https://docs.archway.io). 
