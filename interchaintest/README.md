# Interchain Test

## Chain upgrade test

To run the chain upgrade test locally:


   1. Install Heighliner
   
   ```sh
   git clone https://github.com/strangelove-ventures/heighliner.git
   cd heighliner
   go install
   ```
   
   2. Build the current branch docker image
   
   ```sh
   heighliner build --org archway-network --repo archway --dockerfile cosmos --build-target "make build" --build-env "BUILD_TAGS=muslc" --binaries "build/archwayd" --git-ref <local_branch_name> --tag local
   docker image tag acrechain:local archway:local # There is an issue with heighliner where it wrongly names the docker image
   docker rmi acrechain:local
   ```
   
   3. Build the last release docker image
   
   ```sh
   heighliner build --org archway-network --repo archway --dockerfile cosmos --build-target "make build" --build-env "BUILD_TAGS=muslc" --binaries "build/archwayd" --git-ref v3.0.0 --tag local
   docker image tag acrechain:3.0.0archway:3.0.0
   docker rmi acrechain:3.0.0
   ```
   
   4. Now go to Archway repo root and
  
   ```sh
   cd interchaintest
   go test -v -race -run TestChainUpgrade
   ```