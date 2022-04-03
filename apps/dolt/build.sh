#!/usr/bin/env bash
DOLT_VERSION=$1

set -e

if [ -z "$DOLT_VERSION" ]; then
  echo "$DOLT_VERSION not set"
  exit 1
fi

if [ "$DOLT_VERSION" == "latest" ]; then
  curl -L https://github.com/dolthub/dolt/releases/latest/download/install.sh | bash
else
  curl -L https://github.com/dolthub/dolt/releases/download/v${DOLT_VERSION}/install.sh | bash
fi