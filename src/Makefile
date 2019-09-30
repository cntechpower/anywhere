GOPATH= $(shell dirname `pwd`)
GIT_VERSION = $(shell git rev-parse --abbrev-ref HEAD) $(shell git rev-parse HEAD)
VERSION=$(shell git rev-parse --short HEAD)
default: build
build_server:
	mkdir -p bin/
	GOPATH=${GOPATH} go build -o bin/anywhered anywhere/server/*.go

build_agent:
	mkdir -p bin/
	GOPATH=${GOPATH} go build -o bin/anywhere anywhere/agent/*.go
vet:
	GOPATH=${GOPATH} go vet ./anywhere/...
upload: build
	tar -czvf anywhere-$(VERSION).tar.gz bin/ anywhere/credential/
	tar -czvf anywhere-latest.tar.gz bin/ anywhere/credential/
	curl -T  anywhere-$(VERSION).tar.gz -u ftp:ftp ftp://10.0.0.2/ci/anywhere/
	curl -T  anywhere-latest.tar.gz -u ftp:ftp ftp://10.0.0.2/ci/anywhere/
clean:
	rm -rf bin/
	rm -rf *.tar.gz
build: vet build_server build_agent
