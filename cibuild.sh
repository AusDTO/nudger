#!/bin/bash

set -ex

go test -v
go build -x -o bin/nudger
