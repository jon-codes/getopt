.DEFAULT_GOAL = check
.PHONY: FORCE

CC = /usr/local/musl/bin/musl-gcc
GOFLAGS = -ldflags '-linkmode external -extldflags "-static"'
OUTDIR = bin
TESTGEN_BINARY = $(OUTDIR)/testgen

build-testgen: $(TESTGEN_BINARY)
.PHONY: build-testgen

clean:
	rm -f $(TESTGEN_BINARY)
.PHONY: clean

fmt:
	go fmt ./...
.PHONY: fmt

vet:
	go vet ./...
.PHONY: vet

test:
	go test -v -p=4 ./...
.PHONY: test

check: fmt vet test
.PHONY: check

$(TESTGEN_BINARY): FORCE
	@mkdir -p $(OUTDIR)
	CC=$(CC) go build $(GOFLAGS) -o $@ ./cmd/testgen/main.go

deps: FORCE
	go mod verify
	go mod tidy

all: clean deps
.PHONY: all
