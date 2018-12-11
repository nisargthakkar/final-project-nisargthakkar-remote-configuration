#!/bin/sh
set -E
docker build --rm -t local/config-management-server -f Dockerfile.server .
docker build --rm -t local/config-management-client -f Dockerfile.client .
docker build --rm -t local/config-management-userclient -f Dockerfile.userclient .

docker build --rm -t local/config-management-front-tier -f Dockerfile.front-tier .
docker build --rm -t local/config-management-mid-tier -f Dockerfile.mid-tier .