.PHONY: all build build-web build-go run clean proto test test-backend test-frontend test-coverage lint lint-backend lint-frontend

# Default target
all: build

# Generate Protocol Buffers
proto:
	# Backend Protobuf
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/privutil.proto
	# Frontend Protobuf
	cd web && npm install && \
	protoc --plugin=./node_modules/.bin/protoc-gen-ts_proto \
		--ts_proto_out=src/proto \
		--ts_proto_opt=outputServices=nice-grpc,outputServices=generic-definitions,esModuleInterop=true,env=browser \
		--proto_path=../ proto/privutil.proto

# Build Frontend
build-web:
	cd web && npm install && npm run build

# Build Backend (embeds frontend)
build-go: build-web
	# Ensure the embedded directory exists and is populated
	mkdir -p internal/server/dist
	cp -r web/dist/* internal/server/dist/
	go build -v -o privutil ./cmd/privutil/main.go

# Build everything
build: clean build-go

# Run the application
run: build
	./privutil

# Clean build artifacts
clean:
	rm -f privutil privutil.exe
	rm -rf web/dist/*
	rm -rf internal/server/dist/*
	rm -rf web/node_modules

# Run all tests
test: build test-backend test-frontend

# Run backend tests
test-backend: build
	go test -v -cover ./...

# Run frontend tests (excludes config files and proto)
test-frontend: build
	cd web && npm install && npm run test

# Run tests with coverage reports
test-coverage: test
	@echo "=== Backend Coverage ==="
	go test -coverprofile=coverage.out ./internal/api/...
	go tool cover -func=coverage.out | grep total
	go tool cover -html=coverage.out -o coverage.html
	@echo "\n=== Frontend Coverage ==="
	cd web && npm install && npm run test:coverage

# Run all linters
lint: lint-backend lint-frontend

# Run Go linters
lint-backend: build
	go vet ./...
	go fmt ./...

# Run frontend linters
lint-frontend: build
	cd web && npm install && npm run lint
