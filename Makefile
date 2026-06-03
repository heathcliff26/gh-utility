SHELL := bash

REPOSITORY ?= localhost
CONTAINER_NAME ?= gh-utility
TAG ?= latest

# Build the binary
build:
	hack/build.sh

# Build the container image
image:
	podman build -t $(REPOSITORY)/$(CONTAINER_NAME):$(TAG) .

# Run unit-tests
test:
	go test -v -coverprofile=coverprofile.out ./...

# Update dependencies
update-deps:
	hack/update-deps.sh

# Generate cover profile
coverprofile:
	hack/coverprofile.sh

# Run linter
lint:
	golangci-lint run -v

# Format code
fmt:
	gofmt -s -w ./cmd ./pkg

# Validate that all generated files are up to date
validate:
	hack/validate.sh

# Scan code for vulnerabilities using gosec
gosec:
	gosec ./...

# Clean build artifacts
clean:
	hack/clean.sh

# Show this help message
help:
	@echo "Available targets:"
	@echo ""
	@awk '/^#/{c=substr($$0,3);next}c&&/^[[:alpha:]][[:alnum:]_-]+:/{print substr($$1,1,index($$1,":")),c}1{c=0}' $(MAKEFILE_LIST) | column -s: -t
	@echo ""
	@echo "Run 'make <target>' to execute a specific target."

.PHONY: \
	default \
	build \
	image \
	test \
	update-deps \
	coverprofile \
	lint \
	fmt \
	validate \
	gosec \
	clean \
	help \
	$(NULL)
