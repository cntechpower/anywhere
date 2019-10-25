GOPATH= $(shell dirname `pwd`)
GIT_VERSION = $(shell git rev-parse --abbrev-ref HEAD) $(shell git rev-parse HEAD)
VERSION=$(shell git rev-parse --short HEAD)
default: build
regenerate_key:
	mkdir -p credential/
	rm -rf credential/*
	openssl genrsa -out credential/ca.key 2048
	openssl req -x509 -new -nodes -key credential/ca.key -days 10000 -out credential/ca.crt -subj "/CN=cntechpower_anywhere"
	openssl genrsa -out credential/server.key 2048
	openssl req -new -key credential/server.key -out credential/server.csr -subj "/CN=cntechpower_anywhere"
	openssl x509 -req -in credential/server.csr -CA credential/ca.crt -CAkey credential/ca.key -CAcreateserial -out credential/server.crt -days 365
	openssl genrsa -out credential/client.key 2048
	openssl req -new -key credential/client.key -out credential/client.csr -subj "/CN=cntechpower_anywhere"
	openssl x509 -req -in credential/client.csr -CA credential/ca.crt -CAkey credential/ca.key -CAcreateserial -out credential/client.crt -days 365
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
