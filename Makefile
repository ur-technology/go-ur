# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: gur gur-cross evm all test travis-test-with-coverage xgo clean
.PHONY: gur-linux gur-linux-386 gur-linux-amd64
.PHONY: gur-linux-arm gur-linux-arm-5 gur-linux-arm-6 gur-linux-arm-7 gur-linux-arm64
.PHONY: gur-darwin gur-darwin-386 gur-darwin-amd64
.PHONY: gur-windows gur-windows-386 gur-windows-amd64
.PHONY: gur-android gur-ios

GOBIN = build/bin
GO ?= latest

gur:
	build/env.sh go build -i -v $(shell build/flags.sh) -o $(GOBIN)/gur ./cmd/geth
	@echo "Done building."
	@echo "Run \"$(GOBIN)/gur\" to launch gur."

gur-cross: gur-linux gur-darwin gur-windows gur-android gur-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/gur-*

gur-linux: gur-linux-386 gur-linux-amd64 gur-linux-arm
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/gur-linux-*

gur-linux-386: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --dest=$(GOBIN) --targets=linux/386 -v $(shell build/flags.sh) ./cmd/geth
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/gur-linux-* | grep 386

gur-linux-amd64: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --dest=$(GOBIN) --targets=linux/amd64 -v $(shell build/flags.sh) ./cmd/geth
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gur-linux-* | grep amd64

gur-linux-arm: gur-linux-arm-5 gur-linux-arm-6 gur-linux-arm-7 gur-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/gur-linux-* | grep arm

gur-linux-arm-5: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --dest=$(GOBIN) --targets=linux/arm-5 -v $(shell build/flags.sh) ./cmd/geth
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/gur-linux-* | grep arm-5

gur-linux-arm-6: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --dest=$(GOBIN) --targets=linux/arm-6 -v $(shell build/flags.sh) ./cmd/geth
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/gur-linux-* | grep arm-6

gur-linux-arm-7: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --dest=$(GOBIN) --targets=linux/arm-7 -v $(shell build/flags.sh) ./cmd/geth
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/gur-linux-* | grep arm-7

gur-linux-arm64: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --dest=$(GOBIN) --targets=linux/arm64 -v $(shell build/flags.sh) ./cmd/geth
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/gur-linux-* | grep arm64

gur-darwin: gur-darwin-386 gur-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/gur-darwin-*

gur-darwin-386: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --dest=$(GOBIN) --targets=darwin/386 -v $(shell build/flags.sh) ./cmd/geth
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/gur-darwin-* | grep 386

gur-darwin-amd64: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --dest=$(GOBIN) --targets=darwin/amd64 -v $(shell build/flags.sh) ./cmd/geth
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gur-darwin-* | grep amd64

gur-windows: gur-windows-386 gur-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/gur-windows-*

gur-windows-386: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --dest=$(GOBIN) --targets=windows/386 -v $(shell build/flags.sh) ./cmd/geth
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/gur-windows-* | grep 386

gur-windows-amd64: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --dest=$(GOBIN) --targets=windows/amd64 -v $(shell build/flags.sh) ./cmd/geth
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gur-windows-* | grep amd64

gur-android: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --dest=$(GOBIN) --targets=android-21/aar -v $(shell build/flags.sh) ./cmd/geth
	@echo "Android cross compilation done:"
	@ls -ld $(GOBIN)/gur-android-*

gur-ios: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --dest=$(GOBIN) --targets=ios-7.0/framework -v $(shell build/flags.sh) ./cmd/geth
	@echo "iOS framework cross compilation done:"
	@ls -ld $(GOBIN)/gur-ios-*

evm:
	build/env.sh $(GOROOT)/bin/go install -v $(shell build/flags.sh) ./cmd/evm
	@echo "Done building."
	@echo "Run \"$(GOBIN)/evm to start the evm."

all:
	for cmd in `ls ./cmd/`; do \
		 build/env.sh go build -i -v $(shell build/flags.sh) -o $(GOBIN)/$$cmd ./cmd/$$cmd; \
	done

test: all
	build/env.sh go test ./...

travis-test-with-coverage: all
	build/env.sh go vet ./...
	build/env.sh build/test-global-coverage.sh

xgo:
	build/env.sh go get github.com/karalabe/xgo

clean:
	rm -fr build/_workspace/pkg/ Godeps/_workspace/pkg $(GOBIN)/*
