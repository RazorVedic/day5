#!/bin/bash

# Build script for the Product API

set -e

echo "Building Product API..."

# Variables
APP_NAME="day5"
VERSION=${VERSION:-latest}
REGISTRY=${REGISTRY:-""}

# Build the Go binary
echo "Building Go binary..."
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# Build Docker image
echo "Building Docker image..."
if [ -n "$REGISTRY" ]; then
    IMAGE_TAG="$REGISTRY/$APP_NAME:$VERSION"
else
    IMAGE_TAG="$APP_NAME:$VERSION"
fi

docker build -t $IMAGE_TAG .

echo "Build completed successfully!"
echo "Image: $IMAGE_TAG"

# Optionally push to registry
if [ "$PUSH" = "true" ] && [ -n "$REGISTRY" ]; then
    echo "Pushing image to registry..."
    docker push $IMAGE_TAG
    echo "Image pushed successfully!"
fi
