#!/bin/bash -eux

export cwd=$(pwd)

pushd $cwd/dp-dataset-exporter
  make audit
popd 