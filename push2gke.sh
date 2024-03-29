#!/bin/bash

set -e

IMAGE_TAG=gcr.io/homin-dev/ingress-proxy:$1 
docker buildx build --platform linux/amd64 --build-arg=PROGRAM_VER=$1 -t $IMAGE_TAG .
docker push $IMAGE_TAG

IMAGE_TAG_LATEST=gcr.io/homin-dev/ingress-proxy:latest 
docker tag $IMAGE_TAG $IMAGE_TAG_LATEST
docker push $IMAGE_TAG_LATEST

git tag -a $1 -m "add tag for $1"
git push origin main --tags
