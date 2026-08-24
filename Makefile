build:
	go vet
	go build

install:
	go install -ldflags=-s

.PHONY: build install
