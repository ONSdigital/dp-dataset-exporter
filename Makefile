SHELL=bash

BUILD=build
BUILD_ARCH=$(BUILD)/$(GOOS)-$(GOARCH)
BIN_DIR?=.

export GOOS?=$(shell go env GOOS)
export GOARCH?=$(shell go env GOARCH)

export GRAPH_DRIVER_TYPE?=neptune
export GRAPH_ADDR?=ws://localhost:8182/gremlin

VAULT_ADDR?=http://127.0.0.1:8200
BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)
LDFLAGS=-ldflags "-w -s -X 'main.Version=${VERSION}' -X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GitCommit=$(GIT_COMMIT)'"

VAULT_ADDR?='http://127.0.0.1:8200'
DATABASE_ADDRESS?=bolt://localhost:7687

# The following variables are used to generate a vault token for the app. The reason for declaring variables, is that
# its difficult to move the token code in a Makefile action. Doing so makes the Makefile more difficult to
# read and starts introduction if/else statements.
VAULT_POLICY:="$(shell vault policy write -address=$(VAULT_ADDR) write-psk policy.hcl)"
TOKEN_INFO:="$(shell vault token create -address=$(VAULT_ADDR) -policy=write-psk -period=24h -display-name=dp-dataset-exporter)"
APP_TOKEN:="$(shell echo $(TOKEN_INFO) | awk '{print $$6}')"

generate:
	go generate ./...
build:
	@mkdir -p $(BUILD_ARCH)/$(BIN_DIR)
	go build $(LDFLAGS) -o $(BUILD_ARCH)/$(BIN_DIR)/dp-dataset-exporter cmd/dp-dataset-exporter/main.go

debug acceptance:
	HUMAN_LOG=1 VAULT_TOKEN=$(APP_TOKEN) VAULT_ADDR=$(VAULT_ADDR) go run $(LDFLAGS) -race cmd/dp-dataset-exporter/main.go
test:
	go test -cover -race ./...
vault:
	@echo "$(VAULT_POLICY)"
	@echo "$(TOKEN_INFO)"
	@echo "$(APP_TOKEN)"
.PHONY: build debug test vault acceptance generate
