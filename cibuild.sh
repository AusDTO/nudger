#!/bin/bash

set -x

export GOPATH="$(pwd)/_vendor:$GOPATH"

go test -v
