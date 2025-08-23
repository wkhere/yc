build:
	go vet
	go build

install:
	go install

.PHONY: build install
