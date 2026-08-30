#!/usr/bin/env bash
set -x
TAG=${TAG:-"yourname"}

# export TAG if exists in .env
if [ -f .env ]; then
  export $(cat .env | xargs)
fi

go get -v -t ./...
make deploy TAG=$TAG

make clean
