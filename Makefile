VETARGS?=-all
GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)
BIN_NAME?=open-service-broker-sacloud
CURRENT_VERSION = $(gobump show -r version/)
GO_FILES?=$(shell find . -name '*.go')

BUILD_LDFLAGS = "-s -w \
	  -X github.com/sacloud/open-service-broker-sacloud/version.Revision=`git rev-parse --short HEAD`"

.PHONY: default
default: test build

.PHONY: run
run:
	go run $(CURDIR)/*.go $(ARGS)

.PHONY: clean
clean:
	rm -Rf bin/*

.PHONY: tools
tools:
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/motemen/gobump
	go get -v github.com/alecthomas/gometalinter
	gometalinter --install

.PHONY: build
build: bin/open-service-broker-sacloud

bin/open-service-broker-sacloud: $(GO_FILES)
	GOOS=`go env GOOS` GOARCH=`go env GOARCH` CGO_ENABLED=0 \
                go build \
                    -ldflags $(BUILD_LDFLAGS) \
                    -o bin/open-service-broker-sacloud \
                    *.go

.PHONY: test
test: lint
	TEST_DB= go test ./... $(TESTARGS) -v -timeout=30m -parallel=4 ;

.PHONY: test-db
test-db: lint
	TEST_DB=1 go test ./... $(TESTARGS) -v -timeout=30m -parallel=4 ;

.PHONY: lint
lint: fmt
	gometalinter --vendor --cyclo-over=15 --disable=gas --disable=maligned --deadline=2m ./...
	@echo

.PHONY: fmt
fmt:
	gofmt -s -l -w $(GOFMT_FILES)

.PHONY: bump-patch bump-minor bump-major
bump-patch:
	gobump patch -w version/

bump-minor:
	gobump minor -w version/

bump-major:
	gobump major -w version/
