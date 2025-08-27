# go-foundation-v2

Evangelizing a single opionated standard way to develop software following domain driven design and with its conceptuation done using event modeling. 

## Example Service

This template repo is a mono repo supportig multiple services. It supports one service at present which is example service. In future will be referred by AI models to create other services

- cmd/example is the template service genie would refer.

## Contributing

Open to contribution

## Usage

### Renaming the Service
To rename the service from the default `go-foundation-v2` to your desired service name (e.g., `my-new-service`), run:

```bash
make rename NAME=my-new-service
```
This will:
- Update the module path in `go.mod`, Go source files, and the `Makefile`
- Set the git remote origin URL to `git@github.com:razorpay/my-new-service.git`

Note: After running this command, you'll need to:
1. Create a new repository on GitHub with the same name if not present
2. Run `go mod tidy` to update dependencies
3. Manually update Dockerfiles and GitHub workflow files

### Containerized Build (Recommended for CI)
```
make build
make lint
make test
make proto-generate
```

### Local Development
For faster local development, you can use local Go installation instead of Docker containers:

```
# Build a specific service using local Go installation
make go-build-local BINARY=example
make go-build-local BINARY=example_migration

# Run lint using local golangci-lint if installed
make lint-local
```

Note: If you have more than one service then prefix above commands like: `BINS=payment make build`

## Thanks

"Modern software is built like a tower of blocks; each layer relies on the stability and strength of below."

- Thanks to [upi-switch](https://github.com/razorpay/upi-switch/) from which this was forked and made.
- Thanks to [thockin/go-build-template](https://github.com/thockin/go-build-template).
- Thanks to DDD, event modeling which the repo adopts for design.
- Thanks to Generative AI and paul aider's automation on top of it for helping us move fast: https://github.com/paul-gauthier/aider

"No one ever makes a new technology, they make new combinations of old technolgies."
