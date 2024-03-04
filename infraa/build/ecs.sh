#!/bin/sh
#
# Upload image to ecr

APP_NAME="${APP_NAME:-dashdotdash}"
AWS_PROFILE="${AWS_PROFILE:-default}"

aws ecr get-login-password --region ap-south-1 --profile "${AWS_PROFILE}" | docker login --username AWS --password-stdin "${AWS_ACCOUNT}.dkr.ecr.ap-south-1.amazonaws.com"

# VERSION=latest
if [[ -z "${VERSION}" ]]; then
  VERSION=$(git rev-parse --short=8 HEAD)
fi

DOCKER_BUILD_TAG="${APP_NAME}:${VERSION}"

function build() {
  echo "Building docker image with tag $DOCKER_BUILD_TAG"

  docker build --platform linux/amd64  -t "${DOCKER_BUILD_TAG}" .
}

function push() {
  echo "pushing to dkr.ecr.ap-south-1.amazonaws.com/${DOCKER_BUILD_TAG}"

  docker tag "${DOCKER_BUILD_TAG}" "${AWS_ACCOUNT}.dkr.ecr.ap-south-1.amazonaws.com/${DOCKER_BUILD_TAG}"
  docker push "${AWS_ACCOUNT}.dkr.ecr.ap-south-1.amazonaws.com/${DOCKER_BUILD_TAG}"

# echo "updating ssm params"
# aws ssm put-parameter --name "/talon/server/prod/version" --value "${VERSION}" --type String --overwrite
}

if [[ -z "$1" ]]; then
  build
  push
fi

case "$1" in
  "build")
    build
    ;;
  "push")
    push
    ;;
  "*")
    echo "invalid input"
    ;;
esac
