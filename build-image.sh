#!/bin/sh

set -eu

IMAGE=${1:-xbzu/paopao-ce}
VERSION=${2:-$(git describe --tags --always --dirty)}
COMMIT=$(git rev-parse --short HEAD)
BUILD_DATE=$(git show -s --format=%cI HEAD)

docker buildx build \
  --load \
  --build-arg "VERSION=$VERSION" \
  --build-arg "COMMIT=$COMMIT" \
  --build-arg "BUILD_DATE=$BUILD_DATE" \
  --tag "$IMAGE:$VERSION" \
  --tag "$IMAGE:latest" \
  --file Dockerfile \
  .

echo "Built $IMAGE:$VERSION locally. This command does not push or deploy it."
