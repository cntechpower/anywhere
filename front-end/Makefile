PWD= $(shell pwd)
VERSION=$(shell git rev-parse --short HEAD)
default: build
build:
	rm -rf node_modules
	/usr/local/nodejs/bin/cnpm install
	/usr/local/nodejs/bin/yarn build
upload:
	tar -czvf anywhere-fe-${VERSION}.tar.gz build/
	tar -czvf anywhere-fe-latest.tar.gz build/
	curl -T anywhere-fe-${VERSION}.tar.gz -u ftp:ftp ftp://10.0.0.4/ci/anywhere-fe/
	curl -T anywhere-fe-latest.tar.gz -u ftp:ftp ftp://10.0.0.4/ci/anywhere-fe/
	rm -f anywhere-fe-${VERSION}.tar.gz
	rm -f anywhere-fe-latest.tar.gz
upgrade:
	/usr/local/nodejs/bin/npm-check-updates -u
	/usr/local/nodejs/bin/cnpm install
