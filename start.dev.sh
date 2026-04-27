#!/usr/bin/env bash
set -e

echo "starting DEV environment..."

docker compose -f compose.dev.yml up --build
