.PHONY: tinyjson-install tinyjson-gen test test-unit test-integ build

MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
CONTRACT_DIR := $(abspath $(dir $(MAKEFILE_PATH)))
ROOT_DIR := $(abspath $(dir $(MAKEFILE_PATH))..)

BUILDER_VERSION := "0.5.0"
BUILDER_IMAGE := "cosmwasm/go-optimizer:${BUILDER_VERSION}"

all: build

tinyjson-install:
	@echo "Installing CosmWasm TinyJson"
	go install github.com/CosmWasm/tinyjson/tinyjson

tinyjson-gen:
	@echo "Generating TinyJson files"
	# Using multiple calls since it is easier to comment them out in case one file should be regenerated.
	tinyjson -all -snake_case $(CONTRACT_DIR)/src/types/msg_instantiate.go
	tinyjson -all -snake_case $(CONTRACT_DIR)/src/types/msg_migrate.go
	tinyjson -all -snake_case $(CONTRACT_DIR)/src/types/msg_execute.go
	tinyjson -all -snake_case $(CONTRACT_DIR)/src/types/msg_sudo.go
	tinyjson -all -snake_case $(CONTRACT_DIR)/src/types/msg_query.go
	tinyjson -all -snake_case $(CONTRACT_DIR)/src/types/msg_ibc.go
	tinyjson -all -snake_case $(CONTRACT_DIR)/src/types/types.go
	tinyjson -all -snake_case $(CONTRACT_DIR)/src/state/params.go
	tinyjson -all -snake_case $(CONTRACT_DIR)/src/state/voting.go
	tinyjson -all -snake_case $(CONTRACT_DIR)/src/state/release_stats.go
	tinyjson -all -snake_case $(CONTRACT_DIR)/src/state/withdraw_stats.go
	tinyjson -all -snake_case $(CONTRACT_DIR)/src/state/ibc_stats.go
	tinyjson -all -snake_case $(CONTRACT_DIR)/src/pkg/archway/custom/query.go
	tinyjson -all -snake_case $(CONTRACT_DIR)/src/pkg/archway/custom/msg.go

test: test-unit test-integ

test-unit:
	@echo "Running UNIT tests"
	@go test $(CONTRACT_DIR)/src/...

test-integ:
	@echo "Running integration tests"
	@go test $(CONTRACT_DIR)/integration/...

build:
	@echo "Building the contract using Docker"
	@docker run --rm -e CHECK=1 -e PAGES -v "$(CONTRACT_DIR):/code" ${BUILDER_IMAGE} .
