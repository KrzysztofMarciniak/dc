#!/usr/bin/env bash
set -e

if [ ! -f ".env" ]; then
  echo "missing .env file"
  exit 1
fi

echo "starting PROD environment..."

docker compose -f compose.prod.yml down
docker compose -f compose.prod.yml up --build
