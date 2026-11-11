NPM ?= npm

default: build

.PHONY: build dev test lint clean

build:
	$(NPM) ci
	$(NPM) run build

dev:
	$(NPM) run dev

test:
	CI=true $(NPM) test

lint:
	$(NPM) run lint

clean:
	rm -rf node_modules build
