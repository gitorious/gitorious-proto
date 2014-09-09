.PHONY: test build build-ssh build-http

build: build-ssh build-http

build-ssh:
	cd gitorious-shell && go build

build-http:
	cd gitorious-http-backend && go build

test:
	go test ./...
