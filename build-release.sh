#!/bin/bash
set -euo pipefail

VERSION="$1"
echo "building new version: $VERSION"

GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o qokl .

tar -czvf qokl-linux-amd64-$VERSION.tar.gz qokl

echo "$DOCKER_PASSWORD" | docker login --username "$DOCKER_USERNAME" --password-stdin

docker build -t seapvnk/qokl:$VERSION .
docker tag seapvnk/qokl:$VERSION seapvnk/qokl:latest

docker push seapvnk/qokl:$VERSION
docker push seapvnk/qokl:latest
