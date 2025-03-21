# Variables
BINARY_NAME = gorex
BIN_DIR = bin
SRC_DIR = cmd
PACKAGE = ./...

# Default target: build both binaries
.PHONY: all
all: build

# Build the main binary (entry-point: cmd/main.go)
.PHONY: build-main
build-main:
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(BINARY_NAME) $(SRC_DIR)/main.go
	@echo "✅ Main build complete! Binary located at $(BIN_DIR)/$(BINARY_NAME)"

# Build the benchmark binary (entry-point: benchmark/benchmark.go)
.PHONY: build-benchmark
build-benchmark:
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/benchmark_gorex benchmark/benchmark.go
	@echo "✅ Benchmark build complete! Binary located at $(BIN_DIR)/benchmark_gorex"

# Build both binaries
.PHONY: build
build: build-main build-benchmark

# Run the main executable from the bin directory
.PHONY: run
run: build-main
	@echo "🚀 Running $(BINARY_NAME)..."
	@$(BIN_DIR)/$(BINARY_NAME)
# @echo "🏁 Execution finished!"

# Run the benchmark executable from the bin directory
.PHONY: run-benchmark
run-benchmark: build-benchmark
	@echo "🚀 Running benchmark_gorex..."
	@$(BIN_DIR)/benchmark_gorex
	@echo "🏁 Execution finished!"

# Test all packages
.PHONY: test
test:
	@echo "🧪 Running tests..."
	@go test -v $(PACKAGE)
	@echo "✅ Tests completed!"

# Build and run the main executable
.PHONY: build-run
build-run: build-main
	@echo "🚀 Running $(BINARY_NAME)..."
	@$(BIN_DIR)/$(BINARY_NAME)
	@echo "🏁 Execution finished!"

# Build, test, and run the main executable
.PHONY: build-test-run
build-test-run: build-main test
	@echo "🚀 Running $(BINARY_NAME)..."
	@$(BIN_DIR)/$(BINARY_NAME)
	@echo "🏁 Execution finished!"

# Clean up the bin directory
.PHONY: clean
clean:
	@echo "🧹 Cleaning up..."
	@rm -rf $(BIN_DIR)
	@echo "✅ Cleanup complete!"

# Help message
.PHONY: help
help:
	@echo "📜 Makefile targets:"
	@echo "  make             - Build both main and benchmark binaries (default)"
	@echo "  make build       - Build both main and benchmark binaries"
	@echo "  make build-main  - Build the main binary into $(BIN_DIR)/$(BINARY_NAME)"
	@echo "  make build-benchmark - Build the benchmark binary into $(BIN_DIR)/benchmark_gorex"
	@echo "  make run         - Build and run the main executable 🚀"
	@echo "  make run-benchmark - Build and run the benchmark executable 🚀"
	@echo "  make test        - Run all tests 🧪"
	@echo "  make build-run   - Build and run the main executable 🚀"
	@echo "  make build-test-run - Build, test, and run the main executable 🚀🧪"
	@echo "  make clean       - Remove the bin directory 🧹"
	@echo "  make help        - Show this help message 📜"
