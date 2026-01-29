BIN_ANTI := "./bin/anti-brute-force"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN_ANTI) -ldflags "$(LDFLAGS)" ./cmd/

.PHONY: build