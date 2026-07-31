.PHONY: test build install clean

# Version information
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "unknown")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%d")
LDFLAGS := -X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.buildDate=$(BUILD_DATE)'

# Build the binary
build:
	go build -ldflags "$(LDFLAGS)" -o dynamite ./cmd/dynamite

# Install to GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" ./cmd/dynamite

# Clean built binaries
clean:
	rm -f dynamite

# Run tests
test:
	@go generate ./pkg/... # generate mocks; requires github.com/uber-go/mock
	@go test ./... -count=1 -v
