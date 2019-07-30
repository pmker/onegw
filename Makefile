
.PHONY: all clean
.PHONY: onegw_linux  onegw-linux-amd64 onegw-darwin-amd64
.PHONY: onegw-darwin onegw-darwin-amd64
.PHONY: onegw-windows onegw-windows-386 onegw-windows-amd64
.PHONY: gitversion


GO ?= latest

MAINDIR = onegw
SERVERMAIN = $(shell pwd)/cmd/$(MAINDIR)/main.go
BUILDDIR = $(shell pwd)/build
GOBIN = $(BUILDDIR)/cmd/$(MAINDIR)
ONEGW_VERSION = $(shell cat buildversion)

gitversion:
	@echo "package onegw" > $(shell pwd)/buildversion.go
	@echo "const ONEGW_VERSION = \""$(shell git rev-parse HEAD)"\"" >> $(shell pwd)/buildversion.go
	@echo "const ONEGW_BUILD_VERSION = \""$(ONEGW_VERSION)"\"" >> $(shell pwd)/buildversion.go

onegw:
	@echo "package onegw" > $(shell pwd)/buildversion.go
	@echo "const ONEGW_VERSION = \""$(shell git rev-parse HEAD)"\"" >> $(shell pwd)/buildversion.go
	@echo "const ONEGW_BUILD_VERSION = \""$(ONEGW_VERSION)"\"" >> $(shell pwd)/buildversion.go
	go build -i -o $(GOBIN)/onegw $(SERVERMAIN)
	@echo "Build server done."
	@echo "Run \"$(GOBIN)/onegw\" to start onegw."

all: onegw-windows onegw-darwin  onegw-linux


clean:
	rm -r $(BUILDDIR)/

onegw-linux: onegw-linux-amd64
	@echo "Linux cross compilation done:"


onegw-linux-amd64:
	@echo "version package version is "$(ONEGW_VERSION)"."
	@echo "package onegw" > $(shell pwd)/buildversion.go
	@echo "const ONEGW_VERSION = \""$(shell git rev-parse HEAD)"\"" >> $(shell pwd)/buildversion.go
	@echo "const ONEGW_BUILD_VERSION = \""$(ONEGW_VERSION)"\"" >> $(shell pwd)/buildversion.go
	env GOOS=linux GOARCH=amd64 go build -i -o $(GOBIN)/onegw-$(ONEGW_VERSION)-linux/onegw $(SERVERMAIN)
	@cp  $(shell pwd)/.env $(GOBIN)/onegw-$(ONEGW_VERSION)-linux/.env
	@cp  $(shell pwd)/bootstrap $(GOBIN)/onegw-$(ONEGW_VERSION)-linux/bootstrap
	@echo "Build server done."
	@ls -ld $(GOBIN)/onegw-$(ONEGW_VERSION)-linux/onegw

onegw-darwin:
	@echo "version package version is "$(ONEGW_VERSION)"."
	@echo "package onegw" > $(shell pwd)/buildversion.go
	@echo "const ONEGW_VERSION = \""$(shell git rev-parse HEAD)"\"" >> $(shell pwd)/buildversion.go
	@echo "const ONEGW_BUILD_VERSION = \""$(ONEGW_VERSION)"\"" >> $(shell pwd)/buildversion.go
	env GOOS=darwin GOARCH=amd64 go build -i -o $(GOBIN)/onegw-$(ONEGW_VERSION)-darwin/onegw $(SERVERMAIN)
	@cp  $(shell pwd)/.env $(GOBIN)/onegw-$(ONEGW_VERSION)-darwin/.env
	@cp  $(shell pwd)/bootstrap $(GOBIN)/onegw-$(ONEGW_VERSION)-darwin/bootstrap
	@echo "Build server done."
	@ls -ld $(GOBIN)/onegw-$(ONEGW_VERSION)-darwin/onegw

onegw-windows: onegw-windows-386 onegw-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/onegw-$(ONEGW_VERSION)-windows/onegw-windows-*

onegw-windows-386:
	@echo "package onegw" > $(shell pwd)/buildversion.go
	@echo "const ONEGW_VERSION = \""$(shell git rev-parse HEAD)"\"" >> $(shell pwd)/buildversion.go
	@echo "const ONEGW_BUILD_VERSION = \""$(ONEGW_VERSION)"\"" >> $(shell pwd)/buildversion.go
	env GOOS=windows GOARCH=386 go build -i -o $(GOBIN)/onegw-$(ONEGW_VERSION)-windows/onegw-windows-386.exe $(SERVERMAIN)
	@cp  $(shell pwd)/.env $(GOBIN)/onegw-$(ONEGW_VERSION)-windows/.env
	@echo "Build server done."
	@ls -ld $(GOBIN)/onegw-$(ONEGW_VERSION)-windows/onegw-windows-386.exe

onegw-windows-amd64:
	@echo "package onegw" > $(shell pwd)/buildversion.go
	@echo "const ONEGW_VERSION = \""$(shell git rev-parse HEAD)"\"" >> $(shell pwd)/buildversion.go
	@echo "const ONEGW_BUILD_VERSION = \""$(ONEGW_VERSION)"\"" >> $(shell pwd)/buildversion.go
	env GOOS=windows GOARCH=amd64 go build -i -o $(GOBIN)/onegw-$(ONEGW_VERSION)-windows/onegw-windows-amd64.exe $(SERVERMAIN)
	@cp  $(shell pwd)/.env $(GOBIN)/onegw-$(ONEGW_VERSION)-windows/.env
	@echo "Build server done."ls
	@ls -ld $(GOBIN)/onegw-$(ONEGW_VERSION)-windows/onegw-windows-amd64.exe



