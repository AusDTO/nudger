#!/bin/bash

set -e

docker run                  \
  --rm                      \
  -v "$PWD":/go/src/nudger  \
  -w /go/src/nudger         \
  -t golang:latest          \
  bash -c ./cibuild.sh
