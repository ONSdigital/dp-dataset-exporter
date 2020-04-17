#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-dataset-exporter
  make test
popd