# Default to building for the host
OS ?= $(shell uname)

# Having this will allow CI scripts to build for many OS's and ARCH's
ARCH := $(or $(ARCH),amd64)

# Path to lint tool
GOLINTER ?= golangci-lint
GOFORMATTER ?= gofmt

# Determine binary file name
BIN_NAME := cbuild2cmake
PROG := build/$(BIN_NAME)
ifneq (,$(findstring indows,$(OS)))
    PROG=build/$(BIN_NAME).exe
    OS=windows
else ifneq (,$(findstring Darwin,$(OS)))
    OS=darwin
else
    # Default to Linux
    OS=linux
endif

SOURCES := $(wildcard cmd/cbuild2cmake/*.go) $(wildcard pkg/*/*.go)

all:
	@echo Pick one of:
	@echo $$ make $(PROG)
	@echo $$ make test-all
	@echo $$ make release
	@echo $$ make clean
	@echo $$ make config
	@echo $$ make coverage-report
	@echo
	@echo Build for different OS's and ARCH's by defining these variables. Ex:
	@echo $$ make OS=windows ARCH=amd64 build/$(BIN_NAME).exe
	@echo $$ make OS=darwin  ARCH=amd64 build/$(BIN_NAME)
	@echo
	@echo Run tests
	@echo $$ make test-all
	@echo
	@echo Release a new version of $(BIN_NAME)
	@echo $$ make release
	@echo
	@echo Clean everything
	@echo $$ make clean
	@echo
	@echo Configure local environment
	@echo $$ make config
	@echo
	@echo Generate a report on code-coverage
	@echo $$ make coverage-report

$(PROG): $(SOURCES)
	@echo Building project
	GOOS=$(OS) GOARCH=$(ARCH) go build -ldflags "-X main.version=`git describe 2>/dev/null || echo unknown`" -o $(PROG) ./cmd/cbuild2cmake

run: $(PROG)
	@./$(PROG) $(ARGS) || true

lint:
	$(GOLINTER) run --config=.golangci.yml

format:
	$(GOFORMATTER) -s -w .

format-check:
	mkdir -p build && $(GOFORMATTER) -d . | tee build/format-check.out
	test ! -s build/format-check.out

.PHONY: test release config
test: $(SOURCES)
	mkdir -p build && GOOS=$(OS) GOARCH=$(ARCH) go test $(ARGS) ./... -coverprofile build/cover.out

test-all: format-check coverage-check lint

coverage-report: test
	go tool cover -html=build/cover.out

coverage-check: test
	@echo Checking if test coverage is above 50%
	test `go tool cover -func build/cover.out | tail -1 | awk '{print ($$3 + 0)*10}'` -ge 500

release: test-all $(PROG)
	@./scripts/release

config:
	@echo "Configuring local environment"
	@go version 2>/dev/null || echo "Need Golang: https://golang.org/doc/install"
	@golangci-lint version 2>/dev/null || echo "Need GolangCi-Lint: https://golangci-lint.run/usage/install/#local-installation"

	# Install pre-commit hooks
	cp scripts/pre-commit .git/hooks/pre-commit
clean:
	rm -rf build
