GOPATH= $(shell dirname `pwd`)
GIT_VERSION = $(shell git rev-parse --abbrev-ref HEAD) $(shell git rev-parse HEAD)
VERSION=$(shell git rev-parse --short HEAD)
RPM_VERSION=master
PROJECT_NAME  = anywhere
DOCKER        = $(shell which docker)
DOCKER-COMPOSE        = $(shell which docker-compose)
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
	openssl x509 -req -in credential/server.csr -CA credential/ca.crt -CAkey credential/ca.key -CAcreateserial -out credential/server.crt -days 3650
	openssl genrsa -out credential/client.key 2048
	openssl req -new -key credential/client.key -out credential/client.csr -subj "/CN=cntechpower_anywhere"
	openssl x509 -req -in credential/client.csr -CA credential/ca.crt -CAkey credential/ca.key -CAcreateserial -out credential/client.crt -days 3650
rpc:
	protoc --go_out=plugins=grpc:. anywhere/server/rpc/definitions/*.proto
	protoc --go_out=plugins=grpc:. anywhere/agent/rpc/definitions/*.proto
api:
	GOPATH=${GOPATH} swagger23 generate server -t anywhere/server/restapi/api --exclude-main -f anywhere/server/restapi/definition/anywhere.yml
	GOPATH=${GOPATH} swagger23 generate client -t anywhere/test  -f anywhere/server/restapi/definition/anywhere.yml
build_server:
	mkdir -p bin/
	rm -rf bin/anywhered
	GOPATH=${GOPATH} go build ${LDFLAGS} -o bin/anywhered anywhere/server/*.go

build_agent:
	mkdir -p bin/
	rm -rf bin/anywhere
	GOPATH=${GOPATH} go build ${LDFLAGS} -o bin/anywhere anywhere/agent/*.go
build_agent/arm:
	mkdir -p bin/
	rm -rf bin/anywhere
	GOPATH=${GOPATH} GOARCH=arm64 GOARM=7 go build ${LDFLAGS} -o bin/anywhere anywhere/agent/*.go

build_test_agent:
	mkdir -p bin/
	GOPATH=${GOPATH} go build ${LDFLAGS} -o bin/test anywhere/test/*.go

build_test_image: build ui
	sudo $(DOCKER) build -t anywhered-test-image:latest -f anywhere/test/dockerfiles/Dockerfile.server .
	sudo $(DOCKER) build -t anywhere-test-image:latest -f anywhere/test/dockerfiles/Dockerfile.agent .

docker_test: docker_test_clean build_test_image
	sudo $(DOCKER-COMPOSE) -f anywhere/test/composefiles/docker-compose.yml up -d
	sudo $(DOCKER) exec composefiles_anywhered_1 /usr/local/anywhere/bin/test
	sleep 40
	sudo $(DOCKER) run -t --rm --network composefiles_anywhere_test_net 10.0.0.2:5000/mysql/mysql_client:8.0.19 -h172.90.101.11 -P4444 -proot -e "select @@version"
	sudo $(DOCKER) run -t --rm --network composefiles_anywhere_test_net 10.0.0.2:5000/mysql/mysql_client:5.7.28 -h172.90.101.11 -P4445 -proot -e "select @@version"
	sudo $(DOCKER) run -t --rm --network composefiles_anywhere_test_net --env PASSWD=sshpass 10.0.0.2:5000/centos:sshpass_client sshpass -p sshpass ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -n root@172.90.101.11 -p 4447 /usr/sbin/ip a
	sudo $(DOCKER) run -t --rm --network composefiles_anywhere_test_net 10.0.0.2:5000/cntechpower/busybox:1.31.1-glibc wget -O - http://172.90.101.11:4446
	sudo $(DOCKER) exec -t composefiles_anywhered_1 bash -c "/usr/local/anywhere/bin/anywhered agent list"
	sudo $(DOCKER) exec -t composefiles_anywhered_1 bash -c "/usr/local/anywhere/bin/anywhered proxy list"
	sudo $(DOCKER-COMPOSE) -f anywhere/test/composefiles/docker-compose.yml down
	sudo $(DOCKER) rmi anywhere-test-image:latest
	sudo $(DOCKER) rmi anywhered-test-image:latest

docker_test_clean:
	sudo $(DOCKER-COMPOSE) -f anywhere/test/composefiles/docker-compose.yml down

vet:
	GOPATH=${GOPATH} go vet ./anywhere/...

unittest:
	GOPATH=${GOPATH} go test -count=1 -v  ./anywhere/...

sonar:
	sonar-scanner \
		  -Dsonar.projectKey=anywhere \
		    -Dsonar.sources=. \
		      -Dsonar.host.url=http://10.0.0.2:9001 \
		        -Dsonar.login=a460f47352f57960e3606b827090f32fdb207b7e

upload: upload_x86 upload_arm
upload_arm: build_arm ui
	tar -czf anywhere-$(VERSION)-arm.tar.gz bin/ credential/ static/
	tar -czf anywhere-latest-arm.tar.gz bin/ credential/ static/
	curl -T anywhere-$(VERSION)-arm.tar.gz -u ftp:ftp ftp://10.0.0.2/ci/anywhere/
	curl -T anywhere-latest-arm.tar.gz -u ftp:ftp ftp://10.0.0.2/ci/anywhere/
	rm -rf anywhere-latest-arm.tar.gz
	rm -rf anywhere-$(VERSION).tar.gz
upload_x86: build ui
	tar -czf anywhere-$(VERSION).tar.gz bin/ credential/ static/
	tar -czf anywhere-latest.tar.gz bin/ credential/ static/
	curl -T anywhere-$(VERSION).tar.gz -u ftp:ftp ftp://10.0.0.2/ci/anywhere/
	curl -T anywhere-latest.tar.gz -u ftp:ftp ftp://10.0.0.2/ci/anywhere/
	rm -rf anywhere-$(VERSION).tar.gz
	rm -rf anywhere-latest.tar.gz
upload_release:
	mv anywhere-$(VERSION).tar.gz /var/www/html/
	rm -rf /var/www/html/anywhere-latest.tar.gz && mv anywhere-latest.tar.gz /var/www/html/

clean:
	rm -rf bin/
	rm -rf *.tar.gz
build: vet build_server build_agent build_test_agent
build_arm: vet build_server build_agent/arm build_test_agent

build_release: vet build_server build_agent newkey ui
	tar -czvf anywhere-$(VERSION).tar.gz bin/ credential/ static/
	tar -czvf anywhere-latest.tar.gz bin/ credential/ static/

docker_rpm:
	sudo $(DOCKER) run -v $(shell pwd)/../:/universe --rm $(DOCKER_IMAGE) -c "(mkdir -p /root/rpmbuild/SOURCES >/dev/null 2>&1); cd /root/rpmbuild/SOURCES; (tar zcf ${PROJECT_NAME}.tar.gz /universe --exclude=/universe/pkg --exclude=/universe/bin --transform 's/universe/${PROJECT_NAME}-${RPM_VERSION}/' >/tmp/build.log 2>&1) && (rpmbuild -bb --define 'Version ${RPM_VERSION}' --with qa /universe/src/${PROJECT_NAME}.spec >>/tmp/build.log 2>&1) && (cat /root/rpmbuild/RPMS/x86_64/${PROJECT_NAME}-${RPM_VERSION}-qa.x86_64.rpm) || (cat /tmp/build.log && exit 1)" > ${PROJECT_NAME}-${RPM_VERSION}.x86_64.rpm

upload_rpm:
	curl -T  ${PROJECT_NAME}-${RPM_VERSION}.x86_64.rpm -u ftp:ftp ftp://10.0.0.2/ci/anywhere/rpm/


ui:
	rm -rf static
	wget ftp://ftp:ftp@10.0.0.2/ci/anywhere-fe/anywhere-fe-latest.tar.gz
	tar -xf anywhere-fe-latest.tar.gz
	mv build static
	rm -rf anywhere-fe-latest.tar.gz
