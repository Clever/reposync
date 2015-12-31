include golang.mk
.DEFAULT_GOAL := 
PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLEPKG := github.com/rgarcia/reposync
EXECUTABLE := reposync
VERSION := $(shell cat VERSION)

.PHONY: all build clean test vendor $(PKGS)

all: test build

build:
	go build -o bin/$(EXECUTABLE) $(EXECUTABLEPKG)

vendor: golang-godep-vendor-deps
	$(call golang-godep-vendor,$(PKGS))

test: $(PKGS)

$(PKGS): golang-test-all-deps
	$(call golang-test-all,$@)

run: build
	./bin/$(EXECUTABLE)

BUILDS := \
  build/$(EXECUTABLE)-v$(VERSION)-linux-amd64 \
  build/$(EXECUTABLE)-v$(VERSION)-darwin-amd64
COMPRESSED_BUILDS := $(BUILDS:%=%.tar.gz)
RELEASE_ARTIFACTS := $(COMPRESSED_BUILDS:build/%=release/%)

build/$(EXECUTABLE)-v$(VERSION)-darwin-amd64:
	GOARCH=amd64 GOOS=darwin go build -ldflags "-X main.Version $(VERSION)" -o "$@/$(EXECUTABLE)" $(EXECUTABLEPKG)
build/$(EXECUTABLE)-v$(VERSION)-linux-amd64:
	GOARCH=amd64 GOOS=linux go build -ldflags "-X main.Version $(VERSION)" -o "$@/$(EXECUTABLE)" $(EXECUTABLEPKG)

%.tar.gz: %
	tar -C `dirname $<` -zcvf "$<.tar.gz" `basename $<`

$(RELEASE_ARTIFACTS): release/% : build/%
	mkdir -p release
	cp $< $@

release: $(RELEASE_ARTIFACTS)

$(GOPATH)/bin/gh-release:
	wget https://github.com/progrium/gh-release/releases/download/v2.2.0/gh-release_2.2.0_linux_x86_64.tgz
	tar zxvf gh-release_2.2.0_linux_x86_64.tgz
	mv gh-release $(GOPATH)/bin/gh-release
	rm gh-release_2.2.0_linux_x86_64.tgz

publish: $(GOPATH)/bin/gh-release release
	$(GOPATH)/bin/gh-release create rgarcia/$(EXECUTABLE) $(VERSION) $(shell git rev-parse --abbrev-ref HEAD)

clean:
	rm -rf bin/*
	rm -rf build release
