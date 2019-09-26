GOPATH= $(shell dirname `pwd`)
GIT_VERSION = $(shell git rev-parse --abbrev-ref HEAD) $(shell git rev-parse HEAD)
VERSION=$(shell git rev-parse --short HEAD)
default: build_server
build_server:
	mkdir -p bin/
	GOPATH=${GOPATH} go build -o bin/anywhered anywhere/server/*.go

build_agent:
	mkdir -p bin/
	GOPATH=${GOPATH} go build -o bin/anywhere anywhere/agent/*.go
vet:
	go vet ./anywhere/...