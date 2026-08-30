#!/usr/bin/env bash
set -x
TAG=${TAG:-"your_name_here"} # You can also put it in the git ignored .env

# export TAG if exists in .env
if [ -f .env ]; then
  export $(cat .env | xargs)
fi

go get -v -t ./...
make deploy TAG=$TAG

make clean
