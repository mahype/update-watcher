BINARY_NAME := update-watcher
VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT      := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE        := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS     := -s -w \
    -X github.com/mahype/update-watcher/internal/version.Version=$(VERSION) \
    -X github.com/mahype/update-watcher/internal/version.Commit=$(COMMIT) \
    -X github.com/mahype/update-watcher/internal/version.Date=$(DATE)

.PHONY: build clean test lint fmt vet install snapshot

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) .

install: build
	install -m 0755 bin/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

test:
	go test -v -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run ./...

fmt:
	gofumpt -l -w .

vet:
	go vet ./...

clean:
	rm -rf bin/ dist/ coverage.out

snapshot:
	goreleaser build --snapshot --clean
