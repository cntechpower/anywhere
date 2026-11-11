GO ?= go
DOCKER ?= docker
GO_PKG := ./...
GO_BASE ?= golang:1.22

BIN_DIR := bin
CREDENTIAL_DIR := credential

VERSION := $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X 'main.version=$(GIT_BRANCH) $(VERSION)'"

# Certificate generation. The CN must match the agent --tls-server-name flag and
# the SAN must cover every host/IP used to reach the server.
CERT_CN ?= anywhere.local
CERT_SAN ?= DNS:anywhere.local,DNS:localhost,IP:127.0.0.1
CERT_DAYS ?= 825

default: build

.PHONY: help build build_server build_agent test race vet fmt clean newkey rpc api ui build_ci docker_build

help:
	@echo "targets:"
	@echo "  build          build server and agent binaries into $(BIN_DIR)/"
	@echo "  build_server   build $(BIN_DIR)/anywhered"
	@echo "  build_agent    build $(BIN_DIR)/anywhere"
	@echo "  test           go test ./..."
	@echo "  race           go test -race ./..."
	@echo "  vet            go vet ./..."
	@echo "  fmt            gofmt the hand written Go files"
	@echo "  newkey         regenerate TLS certificates (CA, server, client, http)"
	@echo "  rpc            regenerate protobuf code"
	@echo "  api            regenerate swagger server code"
	@echo "  ui             build the frontend into ./static"
	@echo "  clean          remove build artifacts"

build_server:
	@mkdir -p $(BIN_DIR)
	$(GO) build $(LDFLAGS) -o $(BIN_DIR)/anywhered server/main.go

build_agent:
	@mkdir -p $(BIN_DIR)
	$(GO) build $(LDFLAGS) -o $(BIN_DIR)/anywhere agent/main.go

build: build_server build_agent

test:
	$(GO) test -count=1 $(GO_PKG)

race:
	$(GO) test -count=1 -race $(GO_PKG)

vet:
	$(GO) vet $(GO_PKG)

fmt:
	gofmt -w $$(git ls-files '*.go' | grep -v -e 'api/restapi' -e 'rpc/definitions')

build_ci: vet race build_server build_agent

# Generates a CA and the server/client/http certificates used for mutual TLS.
# The server certificate carries both serverAuth and clientAuth EKUs because the
# local CLI reuses it as a client certificate.
newkey:
	@mkdir -p $(CREDENTIAL_DIR)
	rm -rf $(CREDENTIAL_DIR)/*
	openssl genrsa -out $(CREDENTIAL_DIR)/ca.key 2048
	openssl req -x509 -new -nodes -key $(CREDENTIAL_DIR)/ca.key -days 3650 \
		-out $(CREDENTIAL_DIR)/ca.crt -subj "/CN=anywhere-ca" \
		-addext "basicConstraints=critical,CA:TRUE" \
		-addext "keyUsage=critical,keyCertSign,cRLSign"
	openssl genrsa -out $(CREDENTIAL_DIR)/server.key 2048
	openssl req -new -key $(CREDENTIAL_DIR)/server.key -out $(CREDENTIAL_DIR)/server.csr -subj "/CN=$(CERT_CN)"
	printf 'subjectAltName=%s\nextendedKeyUsage=serverAuth,clientAuth\nkeyUsage=digitalSignature,keyEncipherment\nbasicConstraints=CA:FALSE\n' '$(CERT_SAN)' > $(CREDENTIAL_DIR)/server.ext
	openssl x509 -req -in $(CREDENTIAL_DIR)/server.csr -CA $(CREDENTIAL_DIR)/ca.crt -CAkey $(CREDENTIAL_DIR)/ca.key \
		-CAcreateserial -out $(CREDENTIAL_DIR)/server.crt -days $(CERT_DAYS) -extfile $(CREDENTIAL_DIR)/server.ext
	openssl genrsa -out $(CREDENTIAL_DIR)/client.key 2048
	openssl req -new -key $(CREDENTIAL_DIR)/client.key -out $(CREDENTIAL_DIR)/client.csr -subj "/CN=anywhere-client"
	printf 'extendedKeyUsage=clientAuth\nkeyUsage=digitalSignature,keyEncipherment\nbasicConstraints=CA:FALSE\n' > $(CREDENTIAL_DIR)/client.ext
	openssl x509 -req -in $(CREDENTIAL_DIR)/client.csr -CA $(CREDENTIAL_DIR)/ca.crt -CAkey $(CREDENTIAL_DIR)/ca.key \
		-CAcreateserial -out $(CREDENTIAL_DIR)/client.crt -days $(CERT_DAYS) -extfile $(CREDENTIAL_DIR)/client.ext
	openssl genrsa -out $(CREDENTIAL_DIR)/http.key 2048
	openssl req -new -key $(CREDENTIAL_DIR)/http.key -out $(CREDENTIAL_DIR)/http.csr -subj "/CN=$(CERT_CN)"
	printf 'subjectAltName=%s\nextendedKeyUsage=serverAuth\nkeyUsage=digitalSignature,keyEncipherment\nbasicConstraints=CA:FALSE\n' '$(CERT_SAN)' > $(CREDENTIAL_DIR)/http.ext
	openssl x509 -req -in $(CREDENTIAL_DIR)/http.csr -CA $(CREDENTIAL_DIR)/ca.crt -CAkey $(CREDENTIAL_DIR)/ca.key \
		-CAcreateserial -out $(CREDENTIAL_DIR)/http.crt -days $(CERT_DAYS) -extfile $(CREDENTIAL_DIR)/http.ext
	rm -f $(CREDENTIAL_DIR)/*.csr $(CREDENTIAL_DIR)/*.ext $(CREDENTIAL_DIR)/*.srl
	@echo "certificates written to $(CREDENTIAL_DIR)/"

rpc:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		server/api/rpc/definitions/*.proto agent/rpc/definitions/*.proto

# Generated with go-swagger v0.30.5. The --name flag matters: without it the
# generator derives the API type from the spec title and renames the package.
SWAGGER ?= swagger
api:
	$(SWAGGER) generate server -t server/api/http/api --name AnywhereServer --exclude-main -f server/api/http/definition/anywhere.yml

ui:
	cd front-end && npm ci && npm run build
	rm -rf static
	mv front-end/build static

docker_build:
	$(DOCKER) build -t anywhered:$(VERSION) -f docker-build/Dockerfile.server .
	$(DOCKER) build -t anywhere:$(VERSION) -f docker-build/Dockerfile.agent .

clean:
	rm -rf $(BIN_DIR)
	rm -rf static
	rm -f *.tar.gz
