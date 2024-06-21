#!/bin/bash



echo_info () {
  echo "${blue}"
  echo "$1"
  echo "${reset}"
}

echo_error () {
  echo "${red}"
  echo "$1"
  echo "${reset}"
}

echo_success () {
  echo "${green}"
  echo "$1"
  echo "${reset}"
}

# Set localnet settings
if [[ -f "build/archwayd" ]] ;then
  BINARY=build/archwayd
  # Console log text colour
  red=`tput setaf 9`
  green=`tput setaf 10`
  blue=`tput setaf 12`
  reset=`tput sgr0`
else
  BINARY=archwayd
fi
CHAIN_ID=localnet-1
CHAIN_DIR=./data
VALIDATOR_MNEMONIC="guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host"
DEVELOPER_MNEMONIC="friend excite rough reopen cover wheel spoon convince island path clean monkey play snow number walnut pull lock shoot hurry dream divide concert discover"
USER_MNEMONIC="any giant turtle pioneer frequent frown harvest ancient episode junior vocal rent shrimp icon idle echo suspect clean cage eternal sample post heavy enough"
GENESIS_COINS=1000000000000000000000000000000000000stake

setup_chain () {
  # Stop archwayd if it is already running
  if pgrep -x "$BINARY" >/dev/null; then
      echo_error "Terminating $BINARY..."
      killall archwayd
  fi

  # Remove previous data
  echo_info "Removing previous chain data from $CHAIN_DIR/$CHAIN_ID..."
  rm -rf $CHAIN_DIR/$CHAIN_ID


  # Initialize archwayd with "localnet-1" chain id
  echo_info "Initializing $CHAIN_ID..."
  if $BINARY --home $CHAIN_DIR/$CHAIN_ID init test --chain-id=$CHAIN_ID; then
    echo_success "Successfully initialized $CHAIN_ID"
  else
    echo_error "Failed to initialize $CHAIN_ID"
  fi

  # Modify config for development
  config="$CHAIN_DIR/$CHAIN_ID/config/config.toml"
  if [ "$(uname)" = "Linux" ]; then
   sed -i "s/127.0.0.1/0.0.0.0/g" $config
   sed -i "s/cors_allowed_origins = \[\]/cors_allowed_origins = [\"*\"]/g" $config
  else
    sed -i '' "s/127.0.0.1/0.0.0.0/g" $config
    sed -i '' "s/cors_allowed_origins = \[\]/cors_allowed_origins = [\"*\"]/g" $config
  fi
  # modify genesis params for localnet ease of use
  genesis="$CHAIN_DIR/$CHAIN_ID/config/genesis.json"
  # x/gov params change
  # reduce voting period to 2 minutes
  contents="$(jq '.app_state.gov.params.voting_period = "120s"' $genesis)" && echo "${contents}" >  $genesis
  echo_info "Set x/gov voting period to 120 seconds"
  # reduce expedied voting period to 1 minute
  contents="$(jq '.app_state.gov.params.expedited_voting_period = "60s"' $genesis)" && echo "${contents}" >  $genesis
  echo_info "Set x/gov expedited voting period to 60 seconds"
  # reduce minimum deposit amount to 10stake
  contents="$(jq '.app_state.gov.params.min_deposit[0].amount = "10"' $genesis)" && echo "${contents}" >  $genesis
  echo_info "Set x/gov proposal min deposit amount to 10 stake"
  # reduce deposit period to 20seconds
  contents="$(jq '.app_state.gov.params.max_deposit_period = "20s"' $genesis)" && echo "${contents}" >  $genesis
  echo_info "Set x/gov proposal max deposit period to 20 seconds"


  # Adding users
  echo_info "Adding genesis accounts..."
  echo_info "1. validator"
  echo $VALIDATOR_MNEMONIC | $BINARY --home $CHAIN_DIR/$CHAIN_ID keys add validator --recover --keyring-backend test
  $BINARY --home $CHAIN_DIR/$CHAIN_ID genesis add-genesis-account $($BINARY --home $CHAIN_DIR/$CHAIN_ID keys show validator --keyring-backend test -a) $GENESIS_COINS
  echo_info "2. developer"
  echo $DEVELOPER_MNEMONIC | $BINARY --home $CHAIN_DIR/$CHAIN_ID keys add developer --recover --keyring-backend test
  $BINARY --home $CHAIN_DIR/$CHAIN_ID genesis add-genesis-account $($BINARY --home $CHAIN_DIR/$CHAIN_ID keys show developer --keyring-backend test -a) $GENESIS_COINS
  echo_info "3. user"
  echo $USER_MNEMONIC | $BINARY --home $CHAIN_DIR/$CHAIN_ID keys add user --recover --keyring-backend test
  $BINARY --home $CHAIN_DIR/$CHAIN_ID genesis add-genesis-account $($BINARY --home $CHAIN_DIR/$CHAIN_ID keys show user --keyring-backend test -a) $GENESIS_COINS


  # Creating gentx
  echo_info "Creating gentx for validator..."
  $BINARY --home $CHAIN_DIR/$CHAIN_ID genesis gentx validator 100000000000000000000000stake --chain-id $CHAIN_ID --fees 950000000000000000000stake --keyring-backend test


  # Collecting gentx
  echo_info "Collecting gentx..."
  if $BINARY --home $CHAIN_DIR/$CHAIN_ID genesis collect-gentxs; then
    echo_success "Successfully collected genesis txs into genesis.json"
  else
    echo_error "Failed to collect genesis txs"
  fi


  # Validating genesis
  echo_info "Validating genesis..."
  if $BINARY --home $CHAIN_DIR/$CHAIN_ID genesis validate-genesis; then
    echo_success "Successfully validated genesis"
  else
    echo_error "Failed to validate genesis"
  fi
}

if [ "$1" != "continue" ] ;then
  setup_chain
fi

# Starting chain
echo_info "Starting chain..."
$BINARY --home $CHAIN_DIR/$CHAIN_ID start --minimum-gas-prices 0stake
