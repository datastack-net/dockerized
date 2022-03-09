#!/usr/bin/env bash
DOCKERIZED_ROOT=$(realpath $(dirname "$BASH_SOURCE"))

PWD_CMD="pwd"
case "$(uname -s)" in
   Darwin)
     ;;
   Linux)
     ;;
   CYGWIN*|MINGW32*|MSYS*|MINGW*)
     PWD_CMD="pwd -W"
     ;;
   *)
     ;;
esac

function _dotenv () {
  FILE="$1"
  set -o allexport
  source $FILE
  set +o allexport
}

function _dc() {
#  SERVICE=$1;
#  echo $SERVICE
#  _dotenv "$DOCKERIZED_ROOT/.env"
#  _dotenv "$DOCKERIZED_ROOT/config.env"
  docker-compose -f $DOCKERIZED_ROOT/docker-compose.yml "$@"
}

function dockerized() {
  HOST_PWD=$($PWD_CMD)
   _dc run --rm -w //host -v "$HOST_PWD://host" "$@";
}

function dockerized-shell() {
   dockerized --entrypoint="sh" $1 -c "\$(which bash ash sh zsh dash | head -n 1)"
}
