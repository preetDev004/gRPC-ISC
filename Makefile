# ── grpc-isc Makefile ────────────────────────────────────────────────────────
# Requires: Go 1.22+, buf (https://buf.build/docs/installation), Docker

.PHONY: help gen tidy run build build-all docker-build docker-run docker-stop clean

## help: print this help message
help:
	@echo "Usage: make <target>"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*?##/ { printf "  %-18s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

## gen: generate protobuf + gRPC Go code via buf
gen:
	@bash scripts/gen.sh

## tidy: tidy go modules
tidy:
	go mod tidy

## run: run the server locally (requires gen to have been run)
run:
	go run ./cmd/server

## build: compile the all-in-one local dev binary
build:
	CGO_ENABLED=0 go build -o bin/grpc-isc ./cmd/server

## build-all: compile all service binaries
build-all: gen
	CGO_ENABLED=0 go build -o bin/product-service ./cmd/product-service
	CGO_ENABLED=0 go build -o bin/shipping-service ./cmd/shipping-service
	CGO_ENABLED=0 go build -o bin/order-service ./cmd/order-service
	CGO_ENABLED=0 go build -o bin/gateway ./cmd/gateway

## docker-build: build all Docker images via Compose
docker-build:
	docker compose build

## docker-run: build & start with Docker Compose
docker-run:
	docker compose up --build

## docker-stop: stop and remove containers
docker-stop:
	docker compose down

## clean: remove compiled artifacts
clean:
	rm -rf bin/ gen/
