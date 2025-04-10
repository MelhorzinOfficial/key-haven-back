.PHONY: run build build-linux build-darwin-amd64 build-darwin-arm64 build-windows build-force start check-go-version code-quality unittests unittests-verbose unittests-coverage swag update clean clean-go-cache clean-test-cache compose-linux-up compose-linux-down compose-macos-up compose-macos-down compose-setup-dirs help

help:
	@echo "Available commands:"
	@echo "  make run             - Run the application directly from source"
	@echo "  make build           - Build the application binary (default: linux)"
	@echo "  make build-linux     - Build for Linux with code quality checks"
	@echo "  make build-darwin-amd64 - Build for macOS (Intel) with code quality checks"
	@echo "  make build-darwin-arm64 - Build for macOS (Apple Silicon) with code quality checks"
	@echo "  make build-windows   - Build for Windows with code quality checks"
	@echo "  make build-force     - Build ignoring Go version constraints (use if Go version < 1.24)"
	@echo "  make start           - Run the application directly from source"
	@echo "  make check-go-version - Check if you have the required Go version (1.24+)"
	@echo "  make code-quality    - Run code quality checks"
	@echo "  make unittests       - Run unit tests"
	@echo "  make unittests-verbose - Run unit tests with verbose output"
	@echo "  make unittests-coverage - Run unit tests with coverage"
	@echo "  make swag            - Generate Swagger documentation"
	@echo "  make update          - Update dependencies"
	@echo "  make clean           - Clean cache and test data"
	@echo "  make clean-go-cache  - Clean Go cache"
	@echo "  make clean-test-cache - Clean test cache"
	@echo "  make compose-linux-up - Start docker containers (Linux)"
	@echo "  make compose-linux-down - Stop docker containers (Linux)"
	@echo "  make compose-macos-up - Start docker containers (macOS)"
	@echo "  make compose-macos-down - Stop docker containers (macOS)"
	@echo "  make compose-setup-dirs - Set up docker directories"


# Run the built binary
run: build
	./bin/api

# Build the application (default: linux)
build: build-linux

# Build for Linux
build-linux: check-go-version code-quality
	mkdir -p ./bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/api .
	@echo "Linux build complete: ./bin/api"
	@echo "To install system-wide, run: sudo cp ./bin/api /usr/bin/"

# Build for macOS (Intel)
build-darwin-amd64: check-go-version code-quality
	mkdir -p ./bin
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/api .
	@echo "macOS (Intel) build complete: ./bin/api"
	@echo "To install to your user bin, run: mkdir -p ~/bin && cp ./bin/api ~/bin/"

# Build for macOS (Apple Silicon)
build-darwin-arm64: check-go-version code-quality
	mkdir -p ./bin
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o ./bin/api .
	@echo "macOS (Apple Silicon) build complete: ./bin/api"
	@echo "To install to your user bin, run: mkdir -p ~/bin && cp ./bin/api ~/bin/"

# Build for Windows
build-windows: check-go-version code-quality
	mkdir -p ./bin
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/api.exe .
	@echo "Windows build complete: ./bin/api.exe"

# Force build (ignoring Go version)
build-force:
	@echo "Force building (ignoring Go version requirements)..."
	@echo "Note: This might fail if your Go version is incompatible with dependencies"
	mkdir -p ./bin
	GO111MODULE=on GOPROXY=direct go build -mod=mod -ldflags="-s -w" -o ./bin/api .
	@echo "Force build complete: ./bin/api (version constraints ignored)"



# Golang Dev Startup
start: check-go-version
	# make swag
	go run main.go

# Golang Code Quality Checks

check-go-version:
	@echo "Checking Go version..."
	@go version | grep -q "go1.2[4-9]" || (echo "Error: This project requires Go 1.24 or later. Current version:" && go version && echo "To update Go, visit: https://golang.org/dl/" && exit 1)
	@echo "Go version is compatible with this project"

code-quality: check-go-version
	go mod verify
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	golangci-lint run ./...


# Test Targets

unittests:
	go clean -testcache && go test ./test/unittests/...

unittests-verbose:
	go clean -testcache && go test -v ./test/unittests/...

unittests-coverage:
	go clean -testcache && go test -v -coverpkg=./... -coverprofile=coverage.out ./test/unittests/...
	go tool cover -html=coverage.out


# Swagger Targets

swag:
	swag init -g main.go


# Golang Update Dependencies

update:
	go mod tidy

# Golang Clean Cache

clean: clean-go-cache clean-test-cache
	rm -rf ./bin

clean-go-cache:
	go clean -cache

clean-test-cache:
	go clean -testcache


# Docker Compose Targets

compose-linux-up:
	docker compose -f ./setup/docker-compose.linux.yaml up -d

compose-linux-down:
	docker compose -f ./setup/docker-compose.linux.yaml down

compose-macos-up:
	docker compose -f ./setup/docker-compose.macos.yaml up -d

compose-macos-down:
	docker compose -f ./setup/docker-compose.macos.yaml down

compose-setup-dirs:
	sudo rm -rf ./setup/.docker
	mkdir -p ./setup/.docker/mongodb
	sudo chown -R 1001:1001 ./setup/.docker
	sudo chmod -R 775 ./setup/.docker
