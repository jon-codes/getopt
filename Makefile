.DEFAULT_GOAL = all

GOFLAGS = -ldflags '-linkmode external -extldflags "-static"'
BINDIR = bin
TMPDIR = tmp
TESTGEN_SRC = cmd/testgen/main.go
TESTGEN_BIN = $(BINDIR)/testgen
GOPKG = github.com/jon-codes/getopt

## all: run development tasks (default target)
.PHONY: all
all: deps fmt vet test

.PHONY: check
check: deps-check fmt-check vet test-check

## deps: clean deps
.PHONY: deps
deps:
	go mod tidy -v

.PHONY: deps-check
deps-check:
	go mod tidy -diff
	go mod verify

## fmt: go fmt
.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: fmt-check
fmt-check:
	test -z "$(shell gofmt -l .)"

## vet: go vet
.PHONY: vet
vet:
	go vet ./...

## test: go test
.PHONY: test
test:
	go test ./...

.PHONY: test-check
test-check:
	go test -v ./...

## cover: go test coverage
.PHONY: cover
cover: temp
	go test -v -coverprofile $(TMPDIR)/cover.out $(GOPKG)
	go tool cover -html=$(TMPDIR)/cover.out

## testgen-build: build testgen binary
.PHONY: testgen-build
testgen-build:
	CC=$(shell which musl-gcc) go build $(GOFLAGS) -o $(TESTGEN_BIN) $(TESTGEN_SRC)

## testgen: generate tests
.PHONY: testgen
testgen:
	$(TESTGEN_BIN)

## clean: clean output
.PHONY: clean
clean:
	rm -r $(BINDIR) $(TMPDIR)

temp:
	mkdir -p tmp

## help: display this help
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
