DBG_MAKEFILE ?=
ifeq ($(DBG_MAKEFILE),1)
    $(warning ***** starting Makefile for goal(s) "$(MAKECMDGOALS)")
    $(warning ***** $(shell date))
else
    MAKEFLAGS += -s
endif
MAKEFLAGS += --no-builtin-rules
MAKEFLAGS += --warn-undefined-variables
.SUFFIXES:

BINS ?= example example_migration
REGISTRY ?= c.rzp.io/razorpay
BUILD_IMAGE ?= ${REGISTRY}/rzp-docker-image-inventory-multi-arch:rzp-golden-image-base-golang-1.24-alpine3.21
IMAGE_PREFIX ?= go-foundation-v2-

PROTO_GIT_URL ?= https://github.com/razorpay/proto.git
PROTO_BRANCH := master
PROTO_BUILD_IMAGE  ?= harbor.razorpay.com/razorpay/bufbuild:1.6.0_$(OS)_$(ARCH)
PROTO_ROOT := proto/
RPC_ROOT := rpc/

# use any other value and it will use remote master to checking breaking
BREAKING_REPO_LOCAL ?= local

VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT ?= $(shell git rev-parse HEAD) # till spinakker supports tags
TOKEN_GIT ?= $(shell echo $TOKEN_GIT)
ALL_PLATFORMS := linux/amd64 linux/arm linux/arm64 linux/ppc64le linux/s390x windows/amd64
MOD ?= mod
GOFLAGS ?= -buildvcs=false
# for local
#GOFLAGS ?= -buildvcs=false
HTTP_PROXY ?=
HTTPS_PROXY ?=
OS := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
ARCH := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))
# for local
#OS := linux
#ARCH := amd64

TAG ?= $(VERSION)
BIN_EXTENSION :=
ifeq ($(OS), windows)
  BIN_EXTENSION := .exe
endif
SHELL := /usr/bin/env bash -o errexit -o pipefail -o nounset

# TODO: remove this hack
T ?= # Dummy variable to silence warning

# If you want to build all binaries, see the 'all-build' rule.
# If you want to build all containers, see the 'all-container' rule.
# If you want to build AND push all containers, see the 'all-push' rule.
all: # @HELP builds binaries for one platform ($OS/$ARCH)
all: build

# For the following OS/ARCH expansions, we transform OS/ARCH into OS_ARCH
# because make pattern rules don't match with embedded '/' characters.

build-%:
	$(MAKE) build                         \
	    --no-print-directory              \
	    GOOS=$(firstword $(subst _, ,$*)) \
	    GOARCH=$(lastword $(subst _, ,$*))

container-%:
	$(MAKE) container                     \
	    --no-print-directory              \
	    GOOS=$(firstword $(subst _, ,$*)) \
	    GOARCH=$(lastword $(subst _, ,$*))

push-%:
	$(MAKE) push                          \
	    --no-print-directory              \
	    GOOS=$(firstword $(subst _, ,$*)) \
	    GOARCH=$(lastword $(subst _, ,$*))

all-build: # @HELP builds binaries for all platforms
all-build: $(addprefix build-, $(subst /,_, $(ALL_PLATFORMS)))

all-container: # @HELP builds containers for all platforms
all-container: $(addprefix container-, $(subst /,_, $(ALL_PLATFORMS)))

all-push: # @HELP pushes containers for all platforms to the defined registry
all-push: $(addprefix push-, $(subst /,_, $(ALL_PLATFORMS)))

# The following structure defeats Go's (intentional) behavior to always touch
# result files, even if they have not changed.  This will still run `go` but
# will not trigger further work if nothing has actually changed.
OUTBINS = $(foreach bin,$(BINS),bin/$(OS)_$(ARCH)/$(bin)$(BIN_EXTENSION))

build: $(OUTBINS)
	echo

# Directories that we need to create to build/test.
BUILD_DIRS := bin/$(OS)_$(ARCH)                   \
              bin/tools                           \
              .go/bin/$(OS)_$(ARCH)               \
              .go/bin/$(OS)_$(ARCH)/$(OS)_$(ARCH) \
              .go/cache                           \
              .go/pkg

# Each outbin target is just a facade for the respective stampfile target.
# This `eval` establishes the dependencies for each.
$(foreach outbin,$(OUTBINS),$(eval  \
    $(outbin): .go/$(outbin).stamp  \
))
# This is the target definition for all outbins.
$(OUTBINS):
	true

# Each stampfile target can reference an $(OUTBIN) variable.
$(foreach outbin,$(OUTBINS),$(eval $(strip   \
    .go/$(outbin).stamp: OUTBIN = $(outbin)  \
)))
# This is the target definition for all stampfiles.
# This will build the binary under ./.go and update the real binary iff needed.
STAMPS = $(foreach outbin,$(OUTBINS),.go/$(outbin).stamp)
.PHONY: $(STAMPS)
$(STAMPS): go-build
	echo -ne "binary: $(OUTBIN)  "
	if ! cmp -s .go/$(OUTBIN) $(OUTBIN); then  \
	    mv .go/$(OUTBIN) $(OUTBIN);            \
	    date >$@;                              \
	    echo;                                  \
	else                                       \
	    echo "(cached)";                       \
	fi

# This runs the actual `go build` which updates all binaries.
go-build: | $(BUILD_DIRS)
	echo "# building for $(OS)/$(ARCH)"
	docker run                                                  \
	    -i                                                      \
	    --rm                                                    \
	    --privileged                                            \
	    -u root:root                                            \
	    -v $$(pwd):/src                                         \
	    -w /src                                                 \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin                \
	    -v $${HOME}/.netrc:/.netrc                              \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)  \
	    -v $$(go env GOCACHE):/.cache/go-build                  \
	    -v $$(go env GOMODCACHE):/go/pkg/mod:rw                 \
	    --env HTTP_PROXY=$(HTTP_PROXY)                          \
	    --env HTTPS_PROXY=$(HTTPS_PROXY)                        \
	    --env TOKEN_GIT=$(TOKEN_GIT)                            \
	    $(BUILD_IMAGE)                                          \
	    /bin/sh -c "                                            \
	        apk add --update --no-cache git &&                  \
	        ARCH=$(ARCH)                                        \
	        OS=$(OS)                                            \
	        VERSION=$(VERSION)                                  \
	        MOD=$(MOD)                                          \
	        GOFLAGS=$(GOFLAGS)                                  \
	        ./build/build.sh ./...                              \
	    "

# Example: make shell CMD="-c 'date > datefile'"
shell: # @HELP launches a shell in the containerized build environment
shell: | $(BUILD_DIRS)
	echo "# launching a shell in the containerized build environment"
	docker run                                                  \
	    -ti                                                     \
	    --rm                                                    \
	    --privileged                                            \
	    -u root:root                                           \
	    -v $$(pwd):/src                                         \
	    -w /src                                                 \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin                \
	    -v $${HOME}/.netrc:/.netrc                              \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)  \
	    -v $$(pwd)/.go/cache:/.cache                            \
	    -v $$(pwd)/.go/pkg:/go/pkg                              \
	    --env HTTP_PROXY=$(HTTP_PROXY)                          \
	    --env HTTPS_PROXY=$(HTTPS_PROXY)                        \
	    --env TOKEN_GIT=$(TOKEN_GIT)                            \
	    $(BUILD_IMAGE)                                          \
	    /bin/sh $(CMD)

CONTAINER_DOTFILES = $(foreach bin,$(BINS),.container-$(subst /,_,$(REGISTRY)/$(bin))-$(TAG))

jaeger_docker_build:
	echo "# building and running jaeger image on docker"
	echo "# traces will be visible on http://0.0.0.0:16686/"
	docker run -d --name jaeger \
      -e COLLECTOR_ZIPKIN_HTTP_PORT=9411 \
      -p 5775:5775/udp \
      -p 6831:6831/udp \
      -p 6832:6832/udp \
      -p 5778:5778 \
      -p 16686:16686 \
      -p 14268:14268 \
      -p 14250:14250 \
      -p 9411:9411 \
      jaegertracing/all-in-one:1.18

jaeger_clean:
	echo "# removing jaeger docker containers"
	docker rm -f jaeger

jaeger_refresh: jaeger_clean jaeger_docker_build

# We print the container names here, rather than in CONTAINER_DOTFILES so
# they are always at the end of the output.
container containers: # @HELP builds containers for one platform ($OS/$ARCH)
container containers: $(CONTAINER_DOTFILES)
	for bin in $(BINS); do                           \
	    echo "container: $(REGISTRY)/$(IMAGE_PREFIX)$$bin:$(TAG)";  \
	done
	echo

# Each container-dotfile target can reference a $(BIN) variable.
# This is done in 2 steps to enable target-specific variables.
$(foreach bin,$(BINS),$(eval $(strip                                 \
    .container-$(subst /,_,$(REGISTRY)/$(bin))-$(TAG): BIN = $(bin)  \
)))
$(foreach bin,$(BINS),$(eval                                         \
    .container-$(subst /,_,$(REGISTRY)/$(bin))-$(TAG): bin/$(OS)_$(ARCH)/$(bin)$(BIN_EXTENSION) Dockerfile.in  \
))
# This is the target definition for all container-dotfiles.
# These are used to track build state in hidden files.
$(CONTAINER_DOTFILES):
	echo
	DOCKER_BUILDKIT=1 docker build                   \
	    --no-cache                                   \
	    --build-arg ARG_BIN=$(BIN)$(BIN_EXTENSION)   \
	    --build-arg ARG_OS=$(OS)                     \
	    --build-arg ARG_ARCH=$(ARCH)                 \
	    --secret id=git_token,env=TOKEN_GIT          \
	    -t $(REGISTRY)/$(IMAGE_PREFIX)$(BIN):$(TAG)  \
	    -f Dockerfile.in                             \
	    .
	docker images -q $(REGISTRY)/$(IMAGE_PREFIX)$(BIN):$(TAG) > $@
	echo


push: # @HELP pushes the container for one platform ($OS/$ARCH) to the defined registry
push: container
	for bin in $(BINS); do                     \
		docker push $(REGISTRY)/$(IMAGE_PREFIX)$$bin:$(TAG) && \
		echo "docker tag $(REGISTRY)/$(IMAGE_PREFIX)$$bin:$(TAG) $(REGISTRY)/$(IMAGE_PREFIX)$$bin:$(OS)-$(ARCH)-$(COMMIT)" && \
		docker tag $(REGISTRY)/$(IMAGE_PREFIX)$$bin:$(TAG) $(REGISTRY)/$(IMAGE_PREFIX)$$bin:$(OS)-$(ARCH)-$(COMMIT) && \
		docker push $(REGISTRY)/$(IMAGE_PREFIX)$$bin:$(OS)-$(ARCH)-$(COMMIT); \
	done
	echo

# This depends on github.com/estesp/manifest-tool.
manifest-list: # @HELP builds a manifest list of containers for all platforms
manifest-list: all-push
	pushd tools >/dev/null;                                             \
	  export GOBIN=$$(pwd)/../bin/tools;                                \
	  go install github.com/estesp/manifest-tool/v2/cmd/manifest-tool;  \
	  popd >/dev/null
	for bin in $(BINS); do                                    \
	    platforms=$$(echo $(ALL_PLATFORMS) | sed 's/ /,/g');  \
	    bin/tools/manifest-tool                               \
	        --username=oauth2accesstoken                      \
	        --password=$$(gcloud auth print-access-token)     \
	        push from-args                                    \
	        --platforms "$$platforms"                         \
	        --template $(REGISTRY)/$$bin:$(VERSION)__OS_ARCH  \
	        --target $(REGISTRY)/$$bin:$(VERSION);            \
	done

version: # @HELP outputs the version string
version:
	echo version: $(VERSION)
	echo commit: $(COMMIT)

mocks: # @HELP creates mocks
mocks: | $(BUILD_DIRS)
	docker run                                                   \
	    -i                                                       \
	    --rm                                                     \
	    --privileged                                             \
	    -u root:root                                             \
	    -v $$(pwd):/src                                          \
	    -w /src                                                  \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin                 \
	    -v $${HOME}/.netrc:/.netrc                               \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)   \
	    -v $$(pwd)/.go/cache:/.cache                             \
	    -v $$(pwd)/.go/pkg:/go/pkg                               \
	    --env HTTP_PROXY=$(HTTP_PROXY)                           \
	    --env HTTPS_PROXY=$(HTTPS_PROXY)                         \
	    --env TOKEN_GIT=$(TOKEN_GIT)                             \
	    $(BUILD_IMAGE)                                           \
	    /bin/sh -c "                                             \
	        ARCH=$(ARCH)                                         \
	        OS=$(OS)                                             \
	        VERSION=$(VERSION)                                   \
	        MOD=$(MOD)                                           \
	        GOFLAGS=$(GOFLAGS)                                   \
	        ./build/mocks.sh                                     \
	    "

test: # @HELP runs tests, as defined in ./build/test.sh
test: | $(BUILD_DIRS)
	docker run                                                   \
	    -i                                                       \
	    --rm                                                     \
	    --privileged                                             \
	    -u root:root                                             \
	    -v $$(pwd):/src                                          \
	    -w /src                                                  \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin                 \
	    -v $${HOME}/.netrc:/.netrc                               \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)   \
	    -v $$(pwd)/.go/cache:/.cache                             \
	    -v $$(pwd)/.go/pkg:/go/pkg                               \
	    --env HTTP_PROXY=$(HTTP_PROXY)                           \
	    --env HTTPS_PROXY=$(HTTPS_PROXY)                         \
	    --env TOKEN_GIT=$(TOKEN_GIT)                             \
	    $(BUILD_IMAGE)                                           \
	    /bin/sh -c "                                             \
	        apk add --update --no-cache git &&                   \
	        ARCH=$(ARCH)                                         \
	        OS=$(OS)                                             \
	        VERSION=$(VERSION)                                   \
	        MOD=$(MOD)                                           \
	        GOFLAGS=$(GOFLAGS)                                   \
	        ./build/test.sh $(shell go list ./... | grep -v "e2e")  \
	    "

lint: # @HELP runs golangci-lint
lint:
	@OS=$(OS) ARCH=$(ARCH) ./build/lint.sh lint ./...

fmt: # @HELP runs golangci-lint fmt
fmt:
	@OS=$(OS) ARCH=$(ARCH) ./build/lint.sh fmt ./...

$(BUILD_DIRS):
	mkdir -p $@

clean: # @HELP removes built binaries and temporary files
clean: container-clean bin-clean

container-clean:
	rm -rf .container-* .dockerfile-* .push-*

bin-clean:
	test -d .go && chmod -R u+w .go || true
	rm -rf .go bin

proto-clean:
	rm -rf $(RPC_ROOT)

proto-fetch: # @HELP fetch proto files from remote repo
proto-fetch:
	rm -rf rzpproto
	echo "# fetching proto files from razorpay/proto repo, branch: $(PROTO_BRANCH)"
	rm -rf $(PROTO_ROOT) && \
	mkdir -p $(PROTO_ROOT) && \
	cd $(PROTO_ROOT) && \
	git init --quiet && \
	git config core.sparseCheckout true && \
	cp $(CURDIR)/build/proto_modules .git/info/sparse-checkout && \
	git remote add origin $(PROTO_GIT_URL)  && \
	git fetch origin $(PROTO_BRANCH) --quiet && \
	git checkout $(PROTO_BRANCH) --quiet && \
	cp ../buf.yaml . && \
	cp ../buf.lock .

proto-lint: # @HELP verify if the proto files passes the lint checks
proto-lint:
	echo "# linting proto files under proto/"
	docker run                                                  \
		-i                                                      \
		--rm                                                    \
		--privileged                                            \
		-u root:root                                            \
		-v $$(pwd):/src                                         \
		-w /src                                                 \
		-v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin                \
		-v $${HOME}/.netrc:/.netrc                              \
		-v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)  \
		-v $$(pwd)/.go/cache:/.cache                            \
		-v $$(pwd)/.go/pkg:/go/pkg                              \
		--env TOKEN_GIT=$(TOKEN_GIT)                            \
		$(PROTO_BUILD_IMAGE)                                    \
		/bin/sh -c "buf lint proto"

proto-generate: # @HELP generate code for clients from proto files
proto-generate:
	echo "# generating code under rpc/"
	docker run                                                  \
		-i                                                      \
		--rm                                                    \
		--privileged                                            \
		-u root:root                                            \
		-v $$(pwd):/src                                         \
		-w /src                                                 \
		-v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin                \
		-v $${HOME}/.netrc:/.netrc                              \
		-v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)  \
		-v $$(pwd)/.go/cache:/.cache                            \
		-v $$(pwd)/.go/pkg:/go/pkg                              \
		--env TOKEN_GIT=$(TOKEN_GIT)                            \
		$(PROTO_BUILD_IMAGE)                                    \
		/bin/sh -c "buf generate proto --timeout 10m10s"

generate: # @HELP generates all required code in order, lints it and check breaking
generate: lint proto-generate

help: # @HELP prints this message
help:
	echo "VARIABLES:"
	echo "  BINS = $(BINS)"
	echo "  OS = $(OS)"
	echo "  ARCH = $(ARCH)"
	echo "  MOD = $(MOD)"
	echo "  GOFLAGS = $(GOFLAGS)"
	echo "  REGISTRY = $(REGISTRY)"
	echo
	echo "TARGETS:"
	grep -E '^.*: *# *@HELP' $(MAKEFILE_LIST)     \
	    | awk '                                   \
	        BEGIN {FS = ": *# *@HELP"};           \
	        { printf "  %-30s %s\n", $$1, $$2 };  \
	    '

# Variables for local build
IMAGE_NAME = myapp:latest
BINARY ?= example

# The go-build-local target builds using local Go installation
go-build-local: # @HELP builds binary using local Go installation (usage: make go-build-local BINARY=example)
	@echo "# Building binary using local Go installation"
	@echo "# Building for $(OS)/$(ARCH)"
	@if [ ! -d "./cmd/$(BINARY)" ]; then \
		echo "Error: Directory ./cmd/$(BINARY) does not exist"; \
		exit 1; \
	fi
	mkdir -p bin/$(OS)_$(ARCH) && \
	export CGO_ENABLED=0 && \
	export GOOS=$(OS) && \
	export GOARCH=$(ARCH) && \
	export GOFLAGS="-buildvcs=false" && \
	go build -o bin/$(OS)_$(ARCH)/$(BINARY)$(BIN_EXTENSION) ./cmd/$(BINARY)
	@echo "# Build completed: bin/$(OS)_$(ARCH)/$(BINARY)$(BIN_EXTENSION)"

rename: # @HELP renames the service using build/rename.sh (usage: make rename NAME=new-service-name)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME variable must be set. Usage: make rename NAME=<new_service_name>"; \
		exit 1; \
	fi
	@chmod +x build/rename.sh
	@echo "# Renaming service to $(NAME)..."
	@./build/rename.sh "$(NAME)"
	@echo "# Renaming complete. You may need to run 'go mod tidy'."
