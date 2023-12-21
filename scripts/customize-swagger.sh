#!/usr/bin/env bash

# Converting yaml to json
docker run --rm -v "${PWD}":/workdir mikefarah/yq -p yaml -o json ./docs/static/swagger.yaml > ./docs/static/swagger.json

echo "Customizing swagger files..."

swagger_data=$(jq '.' ./docs/static/swagger.json)

# Adding license
swagger_data=$(echo "${swagger_data}" | jq '.info += {"license":{"name":"Business Source License 1.1","url":"https://github.com/archway-network/archway/blob/main/LICENSE"}}')

# Adding external docs
swagger_data=$(echo "${swagger_data}" | jq '. += {"externalDocs":{"description":"Find out more about Archway","url":"https://docs.archway.io"}}')

# Adding Archway, Cosmos and Wasmd tags info
swagger_data=$(echo "${swagger_data}" | jq '. += {"tags":[{"name":"Archway","description":"Archway Network related endpoints"}, {"name":"Cosmos","description":"Cosmos SDK related endpoints", "externalDocs":{"description":"Find out more", "url":"https://docs.cosmos.network/"}}, {"name":"Cosmwasm","description":"Cosmwasm related endpoints", "externalDocs":{"description":"Find out more", "url":"https://docs.cosmwasm.com/"}}]}')

echo "$swagger_data" > ./docs/static/swagger.json

# Minifying json
jq -c . < ./docs/static/swagger.json > ./docs/static/swagger.min.json

# Cleanup
rm ./docs/static/swagger.yaml
rm ./docs/static/swagger.json