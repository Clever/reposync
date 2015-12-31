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
	GOARCH=amd64 GOOS=darwin go build -ldflags "-X main.Version=$(VERSION)" -o "$@/$(EXECUTABLE)" $(EXECUTABLEPKG)
build/$(EXECUTABLE)-v$(VERSION)-linux-amd64:
	GOARCH=amd64 GOOS=linux go build -ldflags "-X main.Version=$(VERSION)" -o "$@/$(EXECUTABLE)" $(EXECUTABLEPKG)

%.tar.gz: %
	tar -C `dirname $<` -zcvf "$<.tar.gz" `basename $<`

$(RELEASE_ARTIFACTS): release/% : build/%
	mkdir -p release
	cp $< $@

release: $(RELEASE_ARTIFACTS)

$(GOPATH)/bin/github-release:
	# assumes linux dev env
	wget https://github.com/aktau/github-release/releases/download/v0.6.2/linux-amd64-github-release.tar.bz2
	tar xjf linux-amd64-github-release.tar.bz2 
	mv bin/linux/amd64/github-release $(GOPATH)/bin/github-release
	rm -rf gh-release_2.2.0_linux_x86_64.tgz bin/linux

publish: $(GOPATH)/bin/github-release release
	$(GOPATH)/bin/github-release -u rgarcia -r reposync -t $(VERSION) -d $(VERSION)

clean:
	rm -rf bin/*
	rm -rf build release
