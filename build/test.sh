#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

export CGO_ENABLED=0
export GO111MODULE=on
export GOFLAGS="${GOFLAGS:-} -mod=${MOD} -buildvcs=false"
export GOPRIVATE="github.com/razorpay"

if [ ! -f ~/.netrc ]; then
  # Set the SSH URL instead of the HTTPS URL
  git config --global url."https://x-access-token:${TOKEN_GIT}@github.com".insteadOf "https://github.com"
fi

echo "Running tests:"
go test -cover -installsuffix "static" "$@"
echo