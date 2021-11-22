# Archway
This is the core implementation of the Archway protocol leveraging the [Cosmos SDK](https://cosmos.network) & [CosmWasm](https://cosmwasm.com), this allows the protocol to reward not only validators but creators for their contributions to the network.

# Instalation
## From source
### Install golang
Go 1.16 is required for Archway.

if you haven't already, install Golang following the [offical documentation](https://golang.org/doc/install). Make sure your `GOBIN` and `GOPATH` are setup.

### Get the source code
Retrieve the source code from the [official repository](https://github.com/archway-network/archway)

```
git clone https://github.com/archway-network/archway
cd archway
git checkout main
```

### Build
You can build with

```
make install
```

this will install the `archwayd` to your `GOPATH`

## Dockerized

We also provide a docker image to help with test setups. There are two modes to use it

Build: `docker build -t archwaynetwork/archwayd:latest .`

## Documentation
You can check out further information from Archway in our [official documentation](https://docs.archway.io)
