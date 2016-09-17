#!/bin/bash

export GOPATH="$(pwd)/_vendor:$GOPATH"

go test -v
