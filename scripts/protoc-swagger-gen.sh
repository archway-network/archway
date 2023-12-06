#!/usr/bin/env bash

set -eo pipefail

mkdir -p ./tmp-swagger-gen
cd proto

proto_dirs=$(find ./archway -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)

for dir in $proto_dirs; do
  # generate swagger files (filter query files)
  query_file=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))
  if [[ ! -z "$query_file" ]]; then
    buf generate --template buf.gen.swagger.yaml $query_file
  fi
done

cd ..

# Fetching the cosmos-sdk version to use the appropriate swagger file
sdkTemplate="{sdk-version}"
sdkVersion=$(go list -m -f '{{ .Version }}' github.com/cosmos/cosmos-sdk)
sed "s/$sdkTemplate/$sdkVersion/g" ./docs/client/config.json > ./tmp-swagger-gen/config.json

# Fetching the archway version to tag in the swagger doc
archwayTemplate="{archway-version}"
archwayVersion=$(echo $(git describe --tags) | sed 's/^v//')
sed -i "s/$archwayTemplate/$archwayVersion/g" ./tmp-swagger-gen/config.json

# combine swagger files
# uses nodejs package `swagger-combine`.
# all the individual swagger files need to be configured in `config.json` for merging
swagger-combine ./tmp-swagger-gen/config.json -o ./docs/client/swagger.yaml -f yaml --continueOnConflictingPaths true --includeDefinitions true

# clean swagger files
rm -rf ./tmp-swagger-gen