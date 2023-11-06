#!/usr/bin/make -f

PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
LEDGER_ENABLED ?= true
# SDK_PACK := $(shell go list -m github.com/cosmos/cosmos-sdk | sed  's/ /\@/g')
BINDIR ?= $(GOPATH)/bin
SIMAPP = ./app
GORELEASER_CROSS_VERSION = v1.20.6
GORELEASER_VERSION = v1.20.0

# for dockerized protobuf tools
DOCKER := $(shell which docker)
BUF_IMAGE=bufbuild/buf@sha256:9dc5d6645f8f8a2d5aaafc8957fbbb5ea64eada98a84cb09654e8f49d6f73b3e
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(BUF_IMAGE)
HTTPS_GIT := https://github.com/archway-network/archway.git
CURRENT_DIR := $(shell pwd)
SHORT_SHA := $(shell git rev-parse --short HEAD)
LATEST_TAG := $(shell git describe --tags --abbrev=0)

# library versions
LIBWASM_VERSION = $(shell go list -m -f '{{ .Version }}' github.com/CosmWasm/wasmvm)

# Release environment variable
RELEASE ?= false
GORELEASER_SKIP_VALIDATE ?= false

export GO111MODULE = on

# process build tags

build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

ifeq ($(WITH_CLEVELDB),yes)
  build_tags += gcc
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
empty = $(whitespace) $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(empty),$(comma),$(build_tags))

# process linker flags

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=archwayd \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=archwayd \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X github.com/archway-network/archwayd/app.Bech32Prefix=archway \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)"

ifeq ($(WITH_CLEVELDB),yes)
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ifeq ($(LINK_STATICALLY),true)
	ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags_comma_sep)" -ldflags '$(ldflags)' -trimpath

# The below include contains the tools and runsim targets.
include contrib/devtools/Makefile

all: install lint test

build: go.sum
ifeq ($(OS),Windows_NT)
	echo unable to build on windows systems
	exit 1
else
	go build -mod=readonly $(BUILD_FLAGS) -o build/archwayd ./cmd/archwayd
endif

build-all: go.sum
ifeq ($(OS),Windows_NT)
	echo unable to build on windows systems
	exit 1
else
	docker run --rm -v "$(CURDIR)":/code -w /code -e LIBWASM_VERSION=$(LIBWASM_VERSION) ghcr.io/goreleaser/goreleaser:$(GORELEASER_VERSION) build --clean --skip-validate
endif

build-contract-tests-hooks:
ifeq ($(OS),Windows_NT)
	go build -mod=readonly $(BUILD_FLAGS) -o build/contract_tests.exe ./cmd/contract_tests
else
	go build -mod=readonly $(BUILD_FLAGS) -o build/contract_tests ./cmd/contract_tests
endif

install: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/archwayd

########################################
### Tools & dependencies

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify

draw-deps:
	@# requires brew install graphviz or apt-get install graphviz
	go get github.com/RobotsAndPencils/goviz
	@goviz -i ./cmd/archwayd -d 2 | dot -Tpng -o dependency-graph.png

clean:
	rm -rf snapcraft-local.yaml build/

distclean: clean
	rm -rf vendor/

########################################
### Testing


test: test-unit
test-all: check test-race test-cover

test-unit:
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./...

test-race:
	@VERSION=$(VERSION) go test -mod=readonly -race -tags='ledger test_ledger_mock' ./...

test-cover:
	@go test -mod=readonly -timeout 30m -race -coverprofile=coverage.txt -covermode=atomic -tags='ledger test_ledger_mock' ./...

benchmark:
	@go test -mod=readonly -bench=. ./...

test-sim-import-export: runsim
	@echo "Running application import/export simulation. This may take several minutes..."
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) -ExitOnFail 50 5 TestAppImportExport

test-sim-multi-seed-short: runsim
	@echo "Running short multi-seed application simulation. This may take awhile!"
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) -ExitOnFail 50 10 TestFullAppSimulation

###############################################################################
###                                Linting                                  ###
###############################################################################

lint:
	golangci-lint run
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -d -s

format:
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs gofmt -w -s
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs misspell -w
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs goimports -w -local github.com/archway-network/archwayd


###############################################################################
###                                Protobuf                                 ###
###############################################################################
PROTO_BUILDER_IMAGE=ghcr.io/cosmos/proto-builder:0.14.0
PROTO_FORMATTER_IMAGE=tendermintdev/docker-build-proto@sha256:aabcfe2fc19c31c0f198d4cd26393f5e5ca9502d7ea3feafbfe972448fee7cae

proto-all: proto-format proto-lint proto-gen

proto-gen:
	@echo "Generating Protobuf files"
	$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(PROTO_BUILDER_IMAGE) sh ./scripts/protocgen.sh
	./scripts/dontcover.sh ./x/tracking
	./scripts/dontcover.sh ./x/rewards

proto-format:
	@echo "Formatting Protobuf files"
	$(DOCKER) run --rm -v $(CURDIR):/workspace \
	--workdir /workspace $(PROTO_FORMATTER_IMAGE) \
	find ./ -not -path "./third_party/*" -name *.proto -exec clang-format -i {} \;

proto-swagger-gen:
	@echo "Generating Protobuf Swagger files"
	$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(PROTO_BUILDER_IMAGE) sh ./scripts/protoc-swagger-gen.sh
	./scripts/ignite-swagger-gen.sh

proto-lint:
	@$(DOCKER_BUF) lint --error-format=json

proto-check-breaking:
	@$(DOCKER_BUF) breaking --against-input $(HTTPS_GIT)#branch=master

docker-build:
	$(DOCKER) run \
		--rm \
		-e LIBWASM_VERSION=$(LIBWASM_VERSION) \
		-e RELEASE=$(RELEASE) \
		-e GITHUB_TOKEN="$(GITHUB_TOKEN)" \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/github.com/archway-network/archway \
		-w /go/src/github.com/archway-network/archway \
		ghcr.io/goreleaser/goreleaser:$(GORELEASER_VERSION) \
		--clean
		--snapshot

release-dryrun:
	$(DOCKER) run \
		--rm \
		-e LIBWASM_VERSION=$(LIBWASM_VERSION) \
		-e RELEASE=$(RELEASE) \
		-e GITHUB_TOKEN="$(GITHUB_TOKEN)" \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/github.com/archway-network/archway \
		-w /go/src/github.com/archway-network/archway \
		ghcr.io/goreleaser/goreleaser:$(GORELEASER_VERSION) \
		--skip-publish \
		--clean \
		--skip-validate

release:
	$(DOCKER) run \
		--rm \
		-e LIBWASM_VERSION=$(LIBWASM_VERSION) \
		-e RELEASE=$(RELEASE) \
		-e GITHUB_TOKEN="$(GITHUB_TOKEN)" \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/github.com/archway-network/archway \
		-w /go/src/github.com/archway-network/archway \
		ghcr.io/goreleaser/goreleaser:$(GORELEASER_VERSION) \
		--clean \
		--skip-validate=$(GORELEASER_SKIP_VALIDATE)

release-cross:
	$(DOCKER) run \
		--rm \
		-e LIBWASM_VERSION=$(LIBWASM_VERSION) \
		-e RELEASE=$(RELEASE) \
		-e GITHUB_TOKEN="$(GITHUB_TOKEN)" \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/github.com/archway-network/archway \
		-w /go/src/github.com/archway-network/archway \
		ghcr.io/goreleaser/goreleaser-cross:$(GORELEASER_CROSS_VERSION) \
		-f .goreleaser-cross.yaml \
		--clean \
		--skip-validate=$(GORELEASER_SKIP_VALIDATE)

check-vuln-deps:
	go list -json -deps ./... | docker run --rm -i sonatypecommunity/nancy:latest sleuth

.PHONY: all install install-debug \
	go-mod-cache draw-deps clean build format \
	test test-all test-build test-cover test-unit test-race \
	test-sim-import-export \

###############################################################################
###                               Run Localnet                              ###
###############################################################################

# Run localnet in a containerized environment, starts new localnet
localnet:
	TAG=$(LATEST_TAG) docker-compose up

# Continue the stopped containerized localnet, starts the stopped containers
localnet-continue:
	TAG=$(LATEST_TAG) CONTINUE="continue" docker-compose up

# Run a new localnet
run: build
	./scripts/localnet.sh

# Continue the existing localnet
run-continue:
	./scripts/localnet.sh continue

.PHONY: localnet run-continue
