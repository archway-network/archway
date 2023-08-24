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
heighliner build --org archway-network --repo archway --dockerfile cosmos --build-target "make build" --build-env "BUILD_TAGS=muslc" --binaries "build/archwayd" --git-ref <local_branch_name> --tag local
docker image tag acrechain:local archway:local # There is an issue with heighliner where it wrongly names the docker image
docker rmi acrechain:local
```

## IBC conformance test

To run the IBC conformance test locally go to Archway repo root and
  
```sh
cd interchaintest
go test -v -race -run TestGaiaConformance
go test -v -race -run TestOsmosisConformance
```

## Chain upgrade test

To run the chain upgrade test locally, first build the last release docker image. The version should match the value of `initialVersion` in [setup.go](./setup.go)

```sh
heighliner build --org archway-network --repo archway --dockerfile cosmos --build-target "make build" --build-env "BUILD_TAGS=muslc" --binaries "build/archwayd" --git-ref v3.0.0 --tag local
docker image tag acrechain:3.0.0archway:3.0.0
docker rmi acrechain:3.0.0
```
   
Now go to Archway repo root and run
  
```sh
cd interchaintest
go test -v -race -run TestChainUpgrade
```