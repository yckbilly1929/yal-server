BINARY ?= yal-server

test:
	go test ./...

build:
	CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/$(BINARY) .
