#!/bin/bash -eux

cwd=$(pwd)

pushd dp-dataset-exporter
  make build
  cp build/dp-dataset-exporter Dockerfile.concourse ../build
popd
