GIT_VERSION = $(shell git rev-parse --abbrev-ref HEAD) $(shell git rev-parse HEAD)
VERSION = $(shell git rev-parse --short HEAD)
PROJECT_NAME = anywhere
GO_BASE = golang:1.19
DOCKER = $(shell which docker)
DOCKER-COMPOSE = ${DOCKER} compose
LDFLAGS = -ldflags "-X 'main.version=\"${GIT_VERSION}\"'"
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
PWD ?= $(shell pwd)
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
	openssl genrsa -out credential/http.key 2048
	openssl req -new -key credential/http.key -out credential/http.csr -subj "/CN=cntechpower_anywhere"
	openssl x509 -req -in credential/http.csr -CA credential/ca.crt -CAkey credential/ca.key -CAcreateserial -out credential/http.crt -days 3650
	tar -czf anywhere-credential.tar.gz credential/
	curl -T anywhere-credential.tar.gz -u ftp:ftp ftp://10.0.0.4/ci/anywhere/
rpc:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=.  --go-grpc_opt=paths=source_relative server/api/rpc/definitions/*.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=.  --go-grpc_opt=paths=source_relative server/api/rpc/definitions/*.proto agent/rpc/definitions/*.proto
api:
	swagger23 generate server -t server/api/http/api --exclude-main -f server/api/http/definition/anywhere.yml
build_server:
	mkdir -p bin/
	rm -rf bin/anywhered
	sudo $(DOCKER) run --rm -v ${PWD}:/usr/src/myapp -w /usr/src/myapp ${GO_BASE} go build ${LDFLAGS} -o bin/anywhered server/main.go

build_agent:
	mkdir -p bin/
	rm -rf bin/anywhere
	sudo $(DOCKER) run --rm -v ${PWD}:/usr/src/myapp -w /usr/src/myapp ${GO_BASE} go build ${LDFLAGS} -o bin/anywhere agent/main.go

build_docker_image: build ui
	sudo $(DOCKER) build -t anywhered-test-image:latest -f docker-build/Dockerfile.server .
	sudo $(DOCKER) build -t anywhere-test-image:latest -f docker-build/Dockerfile.agent .

upload_docker_img: build ui
	sudo $(DOCKER) build -t 10.0.0.4:5000/cntechpower/${PROJECT_NAME}-agent:${VERSION} -f docker-build/Dockerfile.agent .
	sudo $(DOCKER) push 10.0.0.4:5000/cntechpower/${PROJECT_NAME}-agent:${VERSION}

docker_test: delete_credential docker_test_clean get_credential build_docker_image delete_credential
	sudo $(DOCKER-COMPOSE) -f test/composefiles/docker-compose.yml up -d
	sleep 20 #wait mysql init
	sudo $(DOCKER) run -t --rm --network composefiles_anywhere_test_net 10.0.0.4:5000/mysql/mysql_client:8.0.19 -h172.90.101.11 -P4444 -proot -e "select @@version"
	sudo $(DOCKER) run -t --rm --network composefiles_anywhere_test_net 10.0.0.4:5000/mysql/mysql_client:5.7.28 -h172.90.101.11 -P4445 -proot -e "select @@version"
	#sudo $(DOCKER) run -t --rm --network composefiles_anywhere_test_net --env PASSWD=sshpass 10.0.0.4:5000/centos:sshpass_client sshpass -p sshpass ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -n root@172.90.101.11 -p 4447 /usr/sbin/ip a
	sudo $(DOCKER) run -t --rm --network composefiles_anywhere_test_net 10.0.0.4:5000/cntechpower/busybox:1.31.1-glibc wget -O - http://172.90.101.11:4446
	sudo $(DOCKER) exec -t composefiles-anywhered-1 bash -c "/usr/local/anywhere/bin/anywhered agent list"
	sudo $(DOCKER) logs composefiles-anywhered-1
	sudo $(DOCKER) logs composefiles-anywhere-1-1
	sudo $(DOCKER) logs composefiles-anywhere-2-1
	sudo $(DOCKER) logs composefiles-anywhere-3-1
	sudo $(DOCKER-COMPOSE) -f test/composefiles/docker-compose.yml down
	sudo $(DOCKER) rmi anywhere-test-image:latest
	sudo $(DOCKER) rmi anywhered-test-image:latest

docker_test_clean:
	sudo $(DOCKER-COMPOSE) -f test/composefiles/docker-compose.yml down

vet:
	go vet ./...

unittest:
	 go test -count=1 -v  ./anywhere/...

sonar:
	sudo $(DOCKER) run \
        --rm \
        -e SONAR_HOST_URL="http://sonar.stig.top"  \
        -e SONAR_TOKEN="sqp_2463b064b6d4c937e66fe70734f41725df5215a9" \
        -v "${PWD}:/usr/src" \
        sonarsource/sonar-scanner-cli

upload: delete_credential get_credential upload_x86 upload_docker_img delete_credential

get_credential:
	rm -rf credential
	wget ftp://ftp:ftp@10.0.0.4/ci/anywhere/anywhere-credential.tar.gz
	tar -xf anywhere-credential.tar.gz
	rm -rf anywhere-credential.tar.gz
delete_credential:
	rm -rf credential
upload_x86: build ui
	tar -czf anywhere-$(VERSION).tar.gz bin/ credential/ static/
	tar -czf anywhere-latest.tar.gz bin/ credential/ static/
	curl -T anywhere-$(VERSION).tar.gz -u ftp:ftp ftp://10.0.0.4/ci/anywhere/
	curl -T anywhere-latest.tar.gz -u ftp:ftp ftp://10.0.0.4/ci/anywhere/
	rm -rf anywhere-$(VERSION).tar.gz
	rm -rf anywhere-latest.tar.gz

clean:
	rm -rf bin/
	rm -rf *.tar.gz
build: vet build_server build_agent

build_release: vet build_server build_agent newkey ui
	tar -czvf anywhere-$(VERSION).tar.gz bin/ credential/ static/
	tar -czvf anywhere-latest.tar.gz bin/ credential/ static/

ui:
	rm -rf static
	wget ftp://ftp:ftp@10.0.0.4/ci/anywhere-fe/anywhere-fe-latest.tar.gz
	tar -xf anywhere-fe-latest.tar.gz
	mv build static
	rm -rf anywhere-fe-latest.tar.gz

update_ui:
	rm -rf node_modules
	/usr/local/nodejs/bin/cnpm install
	/usr/local/nodejs/bin/yarn build
	tar -czvf anywhere-fe-${VERSION}.tar.gz build/
	tar -czvf anywhere-fe-latest.tar.gz build/
	curl -T anywhere-fe-${VERSION}.tar.gz -u ftp:ftp ftp://10.0.0.4/ci/anywhere-fe/
	curl -T anywhere-fe-latest.tar.gz -u ftp:ftp ftp://10.0.0.4/ci/anywhere-fe/
	rm -f anywhere-fe-${VERSION}.tar.gz
	rm -f anywhere-fe-latest.tar.gz


build_fe_ci:
	cd front-end
	cd front-end && rm -rf node_modules
	cd front-end && npm install
	cd front-end && yarn build

update_fe_in_repo_ci: build_fe_ci
	rm -rf static
	mv front-end/build static

build_ci: vet build_server build_agent newkey update_fe_in_repo_ci
	tar -czvf anywhere-master.tar.gz bin/ credential/ static/ anywhered.json Makefile

.PHONY: help
help:
	$(warning ---------------------------------------------------------------------------------)
	$(warning Supported Variables And Values:)
	$(warning ---------------------------------------------------------------------------------)
	$(foreach v, $(.VARIABLES), $(if $(filter file,$(origin $(v))), $(info $(v)=$($(v)))))

