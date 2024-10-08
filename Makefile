.DEFAULT_GOAL = all

GOPKG := github.com/jon-codes/getopt

BINDIR := bin
TMPDIR := tmp
OBJDIR := obj

CC := gcc
CFLAGS_BASE := -std=c23 -Werror -Wall -Wextra -Wpedantic -Wno-unused-parameter -Wshadow -Wwrite-strings -Wstrict-prototypes -Wold-style-definition -Wredundant-decls -Wnested-externs -Wmissing-include-dirs -Wjump-misses-init -Wlogical-op
CFLAGS := $(CFLAGS_BASE) -O2 $(TESTGEN_INCL)
CFLAGS_DEBUG = $(CFLAGS_BASE) -g -O0 $(TESTGEN_INCL)

TESTGEN_SRCDIR := testgen
TESTGEN_DATADIR := testdata
TESTGEN_BIN := $(BINDIR)/testgen
TESTGEN_DEBUG_BIN := $(BINDIR)/testgen-debug
TESTGEN_SRCS := $(wildcard $(TESTGEN_SRCDIR)/*.c)
TESTGEN_OBJS := $(patsubst $(TESTGEN_SRCDIR)/%.c,$(OBJDIR)/%.o,$(TESTGEN_SRCS))
TESTGEN_DEBUG_OBJS := $(patsubst $(TESTGEN_SRCDIR)/%.c,$(OBJDIR)/%_debug.o,$(TESTGEN_SRCS))
TESTGEN_INCL := -I$(TESTGEN_SRCDIR) -I/usr/include -ljansson
TESTGEN_INPUT := $(TESTGEN_DATADIR)/cases.json
TESTGEN_OUTPUT := $(TESTGEN_DATADIR)/fixtures.json

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
	go test -count=1 -v ./...

## cover: go test coverage
.PHONY: cover
cover: $(TMPDIR)
	go test -v -coverprofile $(TMPDIR)/cover.out $(GOPKG)
	go tool cover -html=$(TMPDIR)/cover.out

## clean: clean output
.PHONY: clean
clean:
	rm -rf $(BINDIR) $(TMPDIR) $(OBJDIR)

$(TMPDIR) $(BINDIR) $(OBJDIR):
	mkdir -p $@

## testgen: generate test data
testgen: $(TESTGEN_OUTPUT)

## testgen-build: build testgen binary
.PHONY: testgen-build
testgen-build: $(TESTGEN_BIN)

.PHONY: testgen-debug
testgen-debug: $(TESTGEN_DEBUG_BIN)

$(TESTGEN_BIN): $(TESTGEN_OBJS) | $(BINDIR)
	$(CC) $(CFLAGS) -o $@ $(TESTGEN_OBJS)

$(TESTGEN_DEBUG_BIN): $(TESTGEN_DEBUG_OBJS) | $(BINDIR)
	$(CC) $(CFLAGS_DEBUG) -o $@ $(TESTGEN_DEBUG_OBJS)

$(OBJDIR)/%.o: $(TESTGEN_SRCDIR)/%.c | $(OBJDIR)
	$(CC) $(CFLAGS) -c $< -o $@

$(OBJDIR)/%_debug.o: $(TESTGEN_SRCDIR)/%.c | $(OBJDIR)
	$(CC) $(CFLAGS_DEBUG) -c $< -o $@

## valgrind: run testgen with valgrind
.PHONY: valgrind
valgrind: $(TESTGEN_DEBUG_BIN)
	valgrind --leak-check=full $(TESTGEN_DEBUG_BIN) -o $(TESTGEN_OUTPUT) $(TESTGEN_INPUT)

## gdb: run testgen with gdb
.PHONY: gdb
gdb: $(TESTGEN_DEBUG_BIN)
	gdb --args $(TESTGEN_DEBUG_BIN) -o $(TESTGEN_OUTPUT) $(TESTGEN_INPUT)

## help: display this help
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
