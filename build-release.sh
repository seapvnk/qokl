#!/bin/bash
set -euox pipefail

VERSION="$1"
echo "building new version: $VERSION"

# release linux64
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o qokl .
tar -czvf qokl-linux-amd64-$VERSION.tar.gz qokl

if [[ -z "${DOCKER_USERNAME}" || -z "${DOCKER_PASSWORD}" ]]; then
  echo "Error: DOCKER_USERNAME and DOCKER_PASSWORD must be set."
  exit 1
fi

# release docker
echo "$DOCKER_PASSWORD" | docker login --username "$DOCKER_USERNAME" --password-stdin

docker build -t seapvnk/qokl:$VERSION .
docker tag seapvnk/qokl:$VERSION seapvnk/qokl:latest

docker push seapvnk/qokl:$VERSION
docker push seapvnk/qokl:latest
