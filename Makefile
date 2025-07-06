# Define the binary name and path
BINARY_NAME=kdiff
BINARY_PATH=./bin/$(BINARY_NAME)

# Define the main package path
MAIN_PACKAGE=./cmd/kdiff

.PHONY: all build run clean test

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_PATH) $(MAIN_PACKAGE)
	@echo "Build complete. Executable: $(BINARY_PATH)"

run:
	@echo "Running $(BINARY_NAME)..."
	go run $(MAIN_PACKAGE)

clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_PATH)
	rm -rf ./bin/
	go clean -modcache
	@echo "Cleanup complete."

test:
	@echo "Running tests (placeholder)..."
	go test ./...