#!/bin/sh

if [ "$#" -eq 1 ]; then
    GO_ENV=$1
fi

if [ -z "$GO_ENV" ]; then
    echo "Need to pass environment"
    exit -1
fi

docker run --publish 80:8080 -e "GO_ENV=${GO_ENV}"  -d --name test go-ascii-server
