.PHONY: build run generate

build:
	mkdir -p bin
	go build -ldflags "-X main.Version=$(VERSION)" -o bin/gateway ./cmd/gateway

generate:
	cd cmd/gateway && wire
	go mod tidy

run:
	./bin/gateway