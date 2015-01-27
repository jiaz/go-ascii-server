#!/bin/sh

docker run --publish 8080:5555 -d --name test go-ascii-server
