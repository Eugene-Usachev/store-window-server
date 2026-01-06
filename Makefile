GOHOSTOS := $(shell go env GOHOSTOS)
VERSION  := $(shell git describe --tags --always)

SERVICES := gateway

init:
	@if ! command -v protoc >/dev/null 2>&1; then \
		echo "ERROR: protoc is required but not installed."; \
		exit 1; \
	fi

	@if ! command -v protoc-gen-go >/dev/null 2>&1; then \
		go install google.golang.org/protobuf/cmd/protoc-gen-go@latest; \
	fi

	@if ! command -v protoc-gen-go-grpc >/dev/null 2>&1; then \
		go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest; \
	fi

	@if ! command -v protoc-gen-validate >/dev/null 2>&1; then \
		go install github.com/envoyproxy/protoc-gen-validate@latest; \
	fi

	@if ! command -v easyp >/dev/null 2>&1; then \
		go install github.com/easyp-tech/easyp/cmd/easyp@latest; \
	fi

	@if ! command -v wire >/dev/null 2>&1; then \
		go install github.com/google/wire/cmd/wire@latest; \
	fi

proto:
	cd api && easyp generate

.PHONY: init proto build run build-% run-%

build-%: init proto
	@PREFIX="./"; \
	if echo "$(SERVICES)" | grep -wq "$*"; then \
  		PREFIX="./services/"; \
	fi; \
	echo "Generating $*"; \
	$(MAKE) -C $${PREFIX}$* generate; \
	echo "Building $*"; \
	$(MAKE) -C $${PREFIX}$* build VERSION=$(VERSION);

run-%: build-%
	@PREFIX="./"; \
	if echo "$(SERVICES)" | grep -wq "$*"; then \
  		PREFIX="./services/"; \
	fi; \
	$(MAKE) -C $${PREFIX}$* run;

.DEFAULT_GOAL := run