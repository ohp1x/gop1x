BINARY := gop1x
VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags "-X github.com/ohp1x/gop1x/cmd.Version=$(VERSION) -X github.com/ohp1x/gop1x/cmd.Commit=$(COMMIT) -X github.com/ohp1x/gop1x/cmd.Date=$(DATE)"

.PHONY: build test lint vet clean install hooks

build:
	go build $(LDFLAGS) -o $(BINARY) .

test:
	go test -race ./...

vet:
	go vet ./...

lint: vet
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed — install: https://golangci-lint.run/"; \
		exit 1; \
	fi

hooks:
	git config core.hooksPath .githooks
	@echo "git hooks enabled (.githooks/)"

clean:
	rm -f $(BINARY)

install:
	go install $(LDFLAGS) .
