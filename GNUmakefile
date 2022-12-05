ifeq ($(GOPATH),)
	GOPATH:=$(shell go env GOPATH)
endif

.EXPORT_ALL_VARIABLES:
	GO_VERSION=$(GO_VERSION)
	ARCH=$(ARCH)

default: build

tools:
	@echo "==> Installing go <=="
	@sh -c "$(CURDIR)/scripts/tools.sh $(GO_VERSION) $(ARCH)"

fmt:
	@echo "==> Formatting code <=="
	find . -name '*.go' | grep -v vendor | xargs gofmt -s -w

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/fmtcheck.sh'"

build: linux windows

linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-extldflags "-static"' -o bin/tfectl_linux_x86_64

windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags '-extldflags "-static"' -o bin/tfectl_win_x86_64

test:
	@echo "==> Testing <=="
	cd cmd/ && go test -v
