#!/bin/sh

# this script instantiates localnet required genesis files

set -e

echo clearing /root/.app
rm -rf /root/.app
echo initting new chain
# init config files
zoned init zoned-id --chain-id localnet

# create accounts
zoned keys add fd --keyring-backend=test

addr=$(zoned keys show fd -a --keyring-backend=test)
val_addr=$(zoned keys show fd  --keyring-backend=test --bech val -a)

# give the accounts some money
zoned add-genesis-account "$addr" 1000000000000stake --keyring-backend=test

# save configs for the daemon
zoned gentx fd 10000000stake --chain-id localnet --keyring-backend=test

# input genTx to the genesis file
zoned collect-gentxs
# verify genesis file is fine
zoned validate-genesis
echo changing network settings
sed -i 's/127.0.0.1/0.0.0.0/g' /root/.app/config/config.toml

echo test account address: "$addr"
echo test account private key: "$(yes | zoned keys export fd --unsafe --unarmored-hex --keyring-backend=test)"
echo account for --from flag "fd"

echo starting network...
zoned start