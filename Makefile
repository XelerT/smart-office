
BINARY_NAME=server
CMD_PATH=./cmd/api/

# --- Targets ---

# Default target
.DEFAULT_GOAL := help

build: tidy
	echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) $(CMD_PATH)
	echo "Build complete: $(BINARY_NAME)"

run: build
	echo "Running $(BINARY_NAME)... (Requires .env file in this directory)"
	./$(BINARY_NAME)

run-dev: tidy
	echo "Running development server (go run)... (Requires .env file in this directory)"
	go run $(CMD_PATH)

tidy:
	echo "Tidying dependencies..."
	go mod tidy

test: tidy
	echo "Running tests..."
	go test -v ./...

clean:
	echo "Cleaning up build artifacts..."
	rm -f $(BINARY_NAME)
	echo "Cleanup complete."

help:
	echo "Available commands:"
	echo "  make build      Build the application binary ($(BINARY_NAME))"
	echo "  make run        Build and run the application binary"
	echo "  make run-dev    Run the application directly using 'go run' (for development)"
	echo "  make test       Run all tests"
	echo "  make tidy       Tidy Go module dependencies"
	echo "  make clean      Remove the built application binary"
	echo "  make help       Display this help message"

.PHONY: build run run-dev tidy test clean help