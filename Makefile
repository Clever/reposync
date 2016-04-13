include golang.mk
.DEFAULT_GOAL :=
PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLEPKG := github.com/rgarcia/reposync
EXECUTABLE := reposync
VERSION := $(shell cat VERSION)

.PHONY: all build clean test vendor $(PKGS)

$(eval $(call golang-version-check,1.6))

all: test build

build:
	go build -ldflags "-X main.Version=$(VERSION)-dev" -o bin/$(EXECUTABLE) $(EXECUTABLEPKG)

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
	rm -rf linux-amd64-github-release.tar.bz2 bin/linux

export GITHUB_TOKEN ?= $(GITHUB_API_TOKEN)
export GITHUB_USER ?= rgarcia
export GITHUB_REPO ?= $(EXECUTABLE)
publish: $(GOPATH)/bin/github-release release
	$(GOPATH)/bin/github-release release -t $(VERSION) -n $(VERSION) -d $(VERSION)
	$(GOPATH)/bin/github-release upload -t $(VERSION) -n $(EXECUTABLE)-v$(VERSION)-darwin-amd64.tar.gz -f release/$(EXECUTABLE)-v$(VERSION)-darwin-amd64.tar.gz
	$(GOPATH)/bin/github-release upload -t $(VERSION) -n $(EXECUTABLE)-v$(VERSION)-linux-amd64.tar.gz -f release/$(EXECUTABLE)-v$(VERSION)-linux-amd64.tar.gz

clean:
	rm -rf bin/*
	rm -rf build release
