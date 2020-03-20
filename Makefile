SHELL=/bin/bash

GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
GOPATH=$(shell go env GOPATH)


BINPATH=bin/$(GOOS)_$(GOARCH)
BINARY=prometheus-adapter

.PHONY: build
build:
	@echo "--> Building binary ..."
	@mkdir -p $(BINPATH) 
	@go build \
     -o="$(BINPATH)/$(BINARY)" \
     ./cmd/adapter
     
clean:
	rm -rf $(BINPATH)/$(BINARY)   

