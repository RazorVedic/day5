#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

# Check if command argument is provided
if [ $# -eq 0 ]; then
    echo "Usage: $0 <command> [args...]"
    echo "Commands: lint, fmt"
    exit 1
fi

COMMAND="$1"
shift  # Remove the command from arguments

# Validate command
case "$COMMAND" in
    lint|fmt)
        ;;
    *)
        echo "Error: Invalid command '$COMMAND'. Use 'lint' or 'fmt'"
        exit 1
        ;;
esac

# Fixed version of golangci-lint
GOLANGCI_LINT_VERSION="v2.1.6"

# Determine OS and ARCH
OS=${OS:-$(go env GOOS)}
ARCH=${ARCH:-$(go env GOARCH)}

# Binary paths
BIN_DIR="bin/tools"
GOLANGCI_LINT_BIN="${BIN_DIR}/golangci-lint"

echo "Running golangci-lint $COMMAND with version ${GOLANGCI_LINT_VERSION}"

# Create bin directory if it doesn't exist
mkdir -p "${BIN_DIR}"

# Check if golangci-lint exists and is the correct version
if [ -f "${GOLANGCI_LINT_BIN}" ]; then
    CURRENT_VERSION=$(${GOLANGCI_LINT_BIN} --version 2>/dev/null | grep -o '[0-9]\+\.[0-9]\+\.[0-9]\+' | head -1 | sed 's/^/v/' || echo "")
    if [ "${CURRENT_VERSION}" != "${GOLANGCI_LINT_VERSION}" ]; then
        echo "# golangci-lint version mismatch (found: ${CURRENT_VERSION}, expected: ${GOLANGCI_LINT_VERSION})"
        echo "# downloading golangci-lint ${GOLANGCI_LINT_VERSION}..."
        rm -f "${GOLANGCI_LINT_BIN}"
    fi
else
    echo "# downloading golangci-lint ${GOLANGCI_LINT_VERSION}..."
fi

# Download golangci-lint if not present or wrong version
if [ ! -f "${GOLANGCI_LINT_BIN}" ]; then
    DOWNLOAD_URL="https://github.com/golangci/golangci-lint/releases/download/${GOLANGCI_LINT_VERSION}/golangci-lint-${GOLANGCI_LINT_VERSION#v}-${OS}-${ARCH}.tar.gz"
    
    echo "# downloading from: ${DOWNLOAD_URL}"
    
    # Download and extract
    curl -sSfL "${DOWNLOAD_URL}" | tar -xz -C "${BIN_DIR}" --strip-components=1 "golangci-lint-${GOLANGCI_LINT_VERSION#v}-${OS}-${ARCH}/golangci-lint"
    
    # Make executable
    chmod +x "${GOLANGCI_LINT_BIN}"
    
    echo "# golangci-lint ${GOLANGCI_LINT_VERSION} installed successfully"
fi

# Run the appropriate golangci-lint command
case "$COMMAND" in
    lint)
        ${GOLANGCI_LINT_BIN} run --timeout 5m "$@"
        ;;
    fmt)
        ${GOLANGCI_LINT_BIN} fmt "$@"
        ;;
esac

RESULT=$?

if [ $RESULT -ne 0 ]; then
    echo "FAIL"
    echo "------------"
    exit 1
fi
echo "PASS"
