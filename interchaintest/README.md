# Interchain Test

To run the interchain tests, we need to use Heighliner.

To install Heighliner

```sh
git clone https://github.com/strangelove-ventures/heighliner.git
cd heighliner
go install
```

Using Heighliner, build a docker image of your local branch

```sh
heighliner build --org archway-network --repo archway --dockerfile cosmos --build-target "make build" --build-env "BUILD_TAGS=muslc" --binaries "build/archwayd" --git-ref <local_branch_name> --chain archway --tag local
```

## IBC conformance test

To run the IBC conformance test locally go to Archway repo root and
  
```sh
cd interchaintest
go test -v -race -run TestGaiaConformance
```

## Chain upgrade test

To run the chain upgrade test locally, first build the last release docker image. The version should match the value of `initialVersion` in [setup.go](./setup.go)

```sh
heighliner build --org archway-network --repo archway --dockerfile cosmos --build-target "make build" --build-env "BUILD_TAGS=muslc" --binaries "build/archwayd" --git-ref v3.0.0 --chain archway --tag local
```
   
Now go to Archway repo root and run
  
```sh
cd interchaintest
go test -v -race -run TestChainUpgrade
```

## CWICA test

To run the IBC conformance test locally go to Archway repo root and
  
```sh
cd interchaintest
go test -v -race -run TestCWICA
```

The contract binary used for testing is located in the artifacts folder and the source is available at https://github.com/archway-network/test-contracts