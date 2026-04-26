#!/usr/bin/env sh
set -e
export GIN_MODE=debug
echo "<DEV>building..."
go build -o app .
echo "<DEV>running..."
./app
