#!/usr/bin/env bash
DOCKERIZED_ROOT=$(realpath $(dirname "$BASH_SOURCE"))

function dockerized-shell() {
   dockerized --entrypoint="sh" $1 -c "\$(which bash ash sh zsh dash | head -n 1)"
}
