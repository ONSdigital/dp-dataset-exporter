SHELL=bash

BUILD=build
BUILD_ARCH=$(BUILD)/$(GOOS)-$(GOARCH)
BIN_DIR?=.

export GOOS?=$(shell go env GOOS)
export GOARCH?=$(shell go env GOARCH)

build:
	@mkdir -p $(BUILD_ARCH)/$(BIN_DIR)
	go build -o $(BUILD_ARCH)/$(BIN_DIR)/dp-dataset-exporter cmd/dp-dataset-exporter/main.go
debug: build
	HUMAN_LOG=1 go run cmd/dp-dataset-exporter/main.go
test:
	go test -cover $(shell go list ./... | grep -v /vendor/)
.PHONY: build debug test