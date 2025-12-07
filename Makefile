# Define the binary name and path
BINARY_NAME=kdelta
BINARY_PATH=./bin/$(BINARY_NAME)

# Define the main package path
MAIN_PACKAGE=./cmd/kdelta

.PHONY: all build run clean test

all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_PATH) $(MAIN_PACKAGE)
	@echo "Build complete. Executable: $(BINARY_PATH)"

# Run the tool
run: build
	$(BINARY_PATH) help

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_PATH)
	rm -rf ./bin/
	rm -f drift.log
	go clean -modcache
	@echo "Cleanup complete."