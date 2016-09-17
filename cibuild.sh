#!/bin/bash

set -x

export GOPATH="$(pwd)/_vendor:$GOPATH"

go test -v

go build -x

mv nudger $CIRCLE_ARTIFACTS
