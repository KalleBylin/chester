GOCACHE ?= $(CURDIR)/.gocache
BIN_DIR ?= $(CURDIR)/bin

.PHONY: build test

build:
	mkdir -p $(BIN_DIR)
	GOCACHE=$(GOCACHE) go build -mod=vendor -buildvcs=false -o $(BIN_DIR)/chester .

test:
	GOCACHE=$(GOCACHE) go test -mod=vendor ./...
