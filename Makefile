EXTENSION_NAME ?= planning

.PHONY: build
build:
	go build -o gh-$(EXTENSION_NAME) ./...

.PHONY: install
install:
	gh extension install .
