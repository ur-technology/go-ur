# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: gur gur-cross evm all test travis-test-with-coverage xgo clean
.PHONY: gur-linux gur-linux-arm gur-linux-386 gur-linux-amd64
.PHONY: gur-darwin gur-darwin-386 gur-darwin-amd64
.PHONY: gur-windows gur-windows-386 gur-windows-amd64
.PHONY: gur-android gur-android-16 gur-android-21

GOBIN = build/bin

MODE ?= default
GO ?= latest

gur:
	build/env.sh go install -v $(shell build/flags.sh) ./cmd/geth
	@mv $(GOBIN)/geth $(GOBIN)/gur
	@echo "Done building."
	@echo "Run \"$(GOBIN)/gur\" to launch gur."

gur-cross: gur-linux gur-darwin gur-windows gur-android
	@echo "Full cross compilation done:"
	@ls -l $(GOBIN)/gur-*

gur-linux: xgo gur-linux-arm gur-linux-386 gur-linux-amd64
	@echo "Linux cross compilation done:"
	@ls -l $(GOBIN)/gur-linux-*

gur-linux-386: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=linux/386 -v $(shell build/flags.sh) ./cmd/geth
	@mv $(GOBIN)/geth $(GOBIN)/gur
	@echo "Linux 386 cross compilation done:"
	@ls -l $(GOBIN)/gur-linux-* | grep 386

gur-linux-amd64: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=linux/amd64 -v $(shell build/flags.sh) ./cmd/geth
	@mv $(GOBIN)/geth $(GOBIN)/gur
	@echo "Linux amd64 cross compilation done:"
	@ls -l $(GOBIN)/gur-linux-* | grep amd64

gur-linux-arm: gur-linux-arm-5 gur-linux-arm-6 gur-linux-arm-7 gur-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -l $(GOBIN)/gur-linux-* | grep arm

gur-linux-arm-5: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=linux/arm-5 -v $(shell build/flags.sh) ./cmd/geth
	@mv $(GOBIN)/geth $(GOBIN)/gur
	@echo "Linux ARMv5 cross compilation done:"
	@ls -l $(GOBIN)/gur-linux-* | grep arm-5

gur-linux-arm-6: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=linux/arm-6 -v $(shell build/flags.sh) ./cmd/geth
	@mv $(GOBIN)/geth $(GOBIN)/gur
	@echo "Linux ARMv6 cross compilation done:"
	@ls -l $(GOBIN)/gur-linux-* | grep arm-6

gur-linux-arm-7: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=linux/arm-7 -v $(shell build/flags.sh) ./cmd/geth
	@mv $(GOBIN)/geth $(GOBIN)/gur
	@echo "Linux ARMv7 cross compilation done:"
	@ls -l $(GOBIN)/gur-linux-* | grep arm-7

gur-linux-arm64: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=linux/arm64 -v $(shell build/flags.sh) ./cmd/geth
	@mv $(GOBIN)/geth $(GOBIN)/gur
	@echo "Linux ARM64 cross compilation done:"
	@ls -l $(GOBIN)/gur-linux-* | grep arm64

gur-darwin: gur-darwin-386 gur-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -l $(GOBIN)/gur-darwin-*

gur-darwin-386: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=darwin/386 -v $(shell build/flags.sh) ./cmd/geth
	@mv $(GOBIN)/geth $(GOBIN)/gur
	@echo "Darwin 386 cross compilation done:"
	@ls -l $(GOBIN)/gur-darwin-* | grep 386

gur-darwin-amd64: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=darwin/amd64 -v $(shell build/flags.sh) ./cmd/geth
	@mv $(GOBIN)/geth $(GOBIN)/gur
	@echo "Darwin amd64 cross compilation done:"
	@ls -l $(GOBIN)/gur-darwin-* | grep amd64

gur-windows: xgo gur-windows-386 gur-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -l $(GOBIN)/gur-windows-*

gur-windows-386: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=windows/386 -v $(shell build/flags.sh) ./cmd/geth
	@mv $(GOBIN)/geth $(GOBIN)/gur
	@echo "Windows 386 cross compilation done:"
	@ls -l $(GOBIN)/gur-windows-* | grep 386

gur-windows-amd64: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=windows/amd64 -v $(shell build/flags.sh) ./cmd/geth
	@mv $(GOBIN)/geth $(GOBIN)/gur
	@echo "Windows amd64 cross compilation done:"
	@ls -l $(GOBIN)/gur-windows-* | grep amd64

gur-android: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=android/* -v $(shell build/flags.sh) ./cmd/geth
	@mv $(GOBIN)/geth $(GOBIN)/gur
	@echo "Android cross compilation done:"
	@ls -l $(GOBIN)/gur-android-*

gur-ios: gur-ios-arm-7 gur-ios-arm64
	@echo "iOS cross compilation done:"
	@ls -l $(GOBIN)/gur-ios-*

gur-ios-arm-7: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=ios/arm-7 -v $(shell build/flags.sh) ./cmd/geth
	@mv $(GOBIN)/geth $(GOBIN)/gur
	@echo "iOS ARMv7 cross compilation done:"
	@ls -l $(GOBIN)/gur-ios-* | grep arm-7

gur-ios-arm64: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=ios-7.0/arm64 -v $(shell build/flags.sh) ./cmd/geth
	@mv $(GOBIN)/geth $(GOBIN)/gur
	@echo "iOS ARM64 cross compilation done:"
	@ls -l $(GOBIN)/gur-ios-* | grep arm64

evm:
	build/env.sh $(GOROOT)/bin/go install -v $(shell build/flags.sh) ./cmd/evm
	@echo "Done building."
	@echo "Run \"$(GOBIN)/evm to start the evm."

all:
	build/env.sh go install -v $(shell build/flags.sh) ./...

test: all
	build/env.sh go test ./...

travis-test-with-coverage: all
	build/env.sh build/test-global-coverage.sh

xgo:
	build/env.sh go get github.com/karalabe/xgo

clean:
	rm -fr build/_workspace/pkg/ Godeps/_workspace/pkg $(GOBIN)/*
