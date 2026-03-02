GOCACHE ?= $(CURDIR)/.gocache

.PHONY: build test

build:
	GOCACHE=$(GOCACHE) go build ./...

test:
	GOCACHE=$(GOCACHE) go test ./...
