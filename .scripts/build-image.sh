#!/bin/bash
set -eu

VERSION="v$(cd auth && make version)"

if [ -n "${CI_GITHUB_SHA:-}" ]; then
  GITHUB_SHA="$CI_GITHUB_SHA"
fi

GITHUB_SHORT_SHA=${GITHUB_SHA:0:7}
IMAGE_TAG="${VERSION}-${GITHUB_SHORT_SHA}"
if [ -n "${IMAGE_TAG_OVERRIDE:-}" ]; then
  IMAGE_TAG="$IMAGE_TAG_OVERRIDE"
fi

if [ "${GITHUB_REF_NAME:-}" = "main" ] && [ -z "${IMAGE_TAG_OVERRIDE:-}" ]; then
  IMAGE_TAG="$VERSION"
fi

docker build \
  --build-arg GIT_COMMIT_HASH=${GITHUB_SHORT_SHA} \
  -t ghcr.io/roledio/roled-auth:${IMAGE_TAG} \
  ./auth

docker login -u "$GITHUB_ACTOR" -p "$GITHUB_TOKEN" ghcr.io
docker push ghcr.io/roledio/roled-auth:${IMAGE_TAG}

if [ "${PUSH_LATEST:-true}" = "true" ] && [ "${GITHUB_REF_NAME:-}" = "main" ] && [ -z "${IMAGE_TAG_OVERRIDE:-}" ]; then
  docker tag ghcr.io/roledio/roled-auth:${IMAGE_TAG} ghcr.io/roledio/roled-auth:latest
  docker push ghcr.io/roledio/roled-auth:latest
fi
