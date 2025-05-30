# Makefile for hauler

# set shell
SHELL=/bin/bash

# set go variables
GO_FILES=./...
GO_COVERPROFILE=coverage.out

# set build variables
BIN_DIRECTORY=bin
DIST_DIRECTORY=dist

# local build of hauler for current platform
# references/configuration from .goreleaser.yaml
build:
	goreleaser build --clean --snapshot --timeout 60m --single-target

# local build of hauler for all platforms
# references/configuration from .goreleaser.yaml
build-all:
	goreleaser build --clean --snapshot --timeout 60m

# local release of hauler for all platforms
# references/configuration from .goreleaser.yaml
release:
	goreleaser release --clean --snapshot --timeout 60m

# install depedencies
install:
	go mod tidy
	go mod download
	CGO_ENABLED=0 go install ./cmd/...

# format go code
fmt:
	go fmt $(GO_FILES)

# vet go code
vet:
	go vet $(GO_FILES)

# test go code
test:
	go test $(GO_FILES) -cover -race -covermode=atomic -coverprofile=$(GO_COVERPROFILE)

# cleanup artifacts
clean:
	rm -rf $(BIN_DIRECTORY) $(DIST_DIRECTORY) $(GO_COVERPROFILE)
