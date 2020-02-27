GOPATH= $(shell dirname `pwd`)
GIT_VERSION = $(shell git rev-parse --abbrev-ref HEAD) $(shell git rev-parse HEAD)
VERSION=$(shell git rev-parse --short HEAD)
RPM_VERSION=master
PROJECT_NAME  = anywhere
DOCKER        = $(shell which docker)
LDFLAGS       = -ldflags "-X 'main.version=\"${RPM_VERSION}-${GIT_VERSION}\"'"
DOCKER_IMAGE  = 10.0.0.2:5000/actiontech/universe-compiler-go1.11-centos6:v2
default: build
newkey:
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
rpc:
	protoc --go_out=plugins=grpc:. anywhere/server/rpc/definitions/*.proto
api:
	 swagger generate server -t anywhere/server/restapi/api --exclude-main -f anywhere/server/restapi/definition/anywhere.yml
build_server:
	mkdir -p bin/
	GOPATH=${GOPATH} go build ${LDFLAGS} -o bin/anywhered anywhere/server/*.go

build_agent:
	mkdir -p bin/
	GOPATH=${GOPATH} go build ${LDFLAGS} -o bin/anywhere anywhere/agent/*.go
vet:
	GOPATH=${GOPATH} go vet ./anywhere/...
upload: build ui
	tar -czvf anywhere-$(VERSION).tar.gz bin/ credential/ static/
	tar -czvf anywhere-latest.tar.gz bin/ credential/ static/
	curl -T  anywhere-$(VERSION).tar.gz -u ftp:ftp ftp://10.0.0.2/ci/anywhere/
	curl -T  anywhere-latest.tar.gz -u ftp:ftp ftp://10.0.0.2/ci/anywhere/
clean:
	rm -rf bin/
	rm -rf *.tar.gz
build: vet build_server build_agent

docker_rpm:
	sudo $(DOCKER) run -v $(shell pwd)/../:/universe --rm $(DOCKER_IMAGE) -c "(mkdir -p /root/rpmbuild/SOURCES >/dev/null 2>&1); cd /root/rpmbuild/SOURCES; (tar zcf ${PROJECT_NAME}.tar.gz /universe --exclude=/universe/pkg --exclude=/universe/bin --transform 's/universe/${PROJECT_NAME}-${RPM_VERSION}/' >/tmp/build.log 2>&1) && (rpmbuild -bb --define 'Version ${RPM_VERSION}' --with qa /universe/src/${PROJECT_NAME}.spec >>/tmp/build.log 2>&1) && (cat /root/rpmbuild/RPMS/x86_64/${PROJECT_NAME}-${RPM_VERSION}-qa.x86_64.rpm) || (cat /tmp/build.log && exit 1)" > ${PROJECT_NAME}-${RPM_VERSION}.x86_64.rpm

upload_rpm:
	curl -T  ${PROJECT_NAME}-${RPM_VERSION}.x86_64.rpm -u ftp:ftp ftp://10.0.0.2/ci/anywhere/rpm/


ui:
	rm -rf static
	wget ftp://ftp:ftp@10.0.0.2/ci/react/react-latest.tar.gz
	tar -xf react-latest.tar.gz
	mv build static
	rm -rf react-latest.tar.gz
