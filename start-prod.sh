#!/usr/bin/env sh
set -e
export GIN_MODE=release
echo "<PROD>building..."
go build -o app .
echo "<PROD>running..."
./app
