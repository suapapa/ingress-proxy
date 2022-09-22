#!/bin/bash
# git tag -a $1 -m "add tag for $1"
# git push --tags

IMAGE_TAG=gcr.io/homin-dev/ingress-proxy:$1 
docker buildx build --platform linux/amd64 -t $IMAGE_TAG .
docker push $IMAGE_TAG
