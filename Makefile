BINARY := gop1x
VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags "-X github.com/ohp1x/gop1x/cmd.Version=$(VERSION) -X github.com/ohp1x/gop1x/cmd.Commit=$(COMMIT) -X github.com/ohp1x/gop1x/cmd.Date=$(DATE)"

.PHONY: build test lint clean install

build:
	go build $(LDFLAGS) -o $(BINARY) .

test:
	go test -race ./...

lint:
	go vet ./...

clean:
	rm -f $(BINARY)

install:
	go install $(LDFLAGS) .
