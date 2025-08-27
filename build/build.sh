#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

if [ -z "${OS:-}" ]; then
    echo "OS must be set"
    exit 1
fi
if [ -z "${ARCH:-}" ]; then
    echo "ARCH must be set"
    exit 1
fi
if [ -z "${VERSION:-}" ]; then
    echo "VERSION must be set"
    exit 1
fi

export CGO_ENABLED=0
export GOARCH="${ARCH}"
export GOOS="${OS}"
export GO111MODULE=on
export GOFLAGS="${GOFLAGS:-} -mod=${MOD} -buildvcs=false"
export GOPRIVATE="github.com/razorpay"

if [ ! -f ~/.netrc ]; then
  # Set the SSH URL instead of the HTTPS URL
  git config --global url."https://x-access-token:${TOKEN_GIT}@github.com".insteadOf "https://github.com"
fi
go install                                                      \
    -installsuffix "static"                                     \
    -ldflags "-X $(go list -m)/pkg/version.Version=${VERSION}"  \
    "$@"