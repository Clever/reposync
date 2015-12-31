include golang.mk
.DEFAULT_GOAL := 
PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLEPKG := github.com/rgarcia/abridgedabridgedsummary
EXECUTABLE := abridgedabridgedsummary
.PHONY: all build clean test vendor $(PKGS)

all: test build

build:
	go build -o bin/$(EXECUTABLE) $(EXECUTABLEPKG)

vendor: golang-godep-vendor-deps
	$(call golang-godep-vendor,$(PKGS))

test: $(PKGS)

$(PKGS): golang-test-all-deps
	$(call golang-test-all,$@)

clean:
	rm bin/*

run: build
	./bin/$(EXECUTABLE)
