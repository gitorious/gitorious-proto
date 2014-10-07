.PHONY: test build build-ssh build-http

build: test build-ssh build-http

deps:
	go get -d -v ./...

test:
	go test ./...

build-ssh:
	cd gitorious-shell && go build

build-http:
	cd gitorious-http-backend && go build

build-ssh-linux:
	cd gitorious-shell && gox -osarch=linux/amd64

build-http-linux:
	cd gitorious-http-backend && gox -osarch=linux/amd64
