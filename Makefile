.PHONY: build

build:
	CGO_ENABLED=0 go build -ldflags "-s -w" -o builds/gochat ./cmd/gochat
