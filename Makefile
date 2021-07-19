include golang.mk
.DEFAULT_GOAL :=
PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLEPKG := github.com/Clever/reposync
EXECUTABLE := reposync
VERSION := $(shell cat VERSION)

.PHONY: all build clean test vendor release $(PKGS)

$(eval $(call golang-version-check,1.16))

all: test build

install_deps:
	go mod vendor

build:
	go build -ldflags "-X main.Version=$(VERSION)" -o bin/$(EXECUTABLE) $(EXECUTABLEPKG)

test: $(PKGS)

$(PKGS): golang-test-all-deps
	$(call golang-test-all,$@)

run: build
	./bin/$(EXECUTABLE)

release:
	@GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" \
		-o="$@/$(EXECUTABLE)-$(VERSION)-linux-amd64"
	@GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" \
		-o="$@/$(EXECUTABLE)-$(VERSION)-darwin-amd64"

clean:
	rm -rf bin/*
