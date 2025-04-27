# Variables
BINARY_NAME = zvm
LIB_NAME = libzscript.so
BIN_DIR = bin
SRC_DIR = cmd
PACKAGE = ./...
SAMPLE_DIR = samples/lib
RUST_DIR = $(SAMPLE_DIR)/rust
RUST_BINARY = sample_rust

# Compiler and linker flags
CC = gcc
CXX = g++
CFLAGS = -I$(BIN_DIR) -L$(BIN_DIR) -lzscript
CXXFLAGS = -I$(BIN_DIR) -L$(BIN_DIR) -lzscript
GOFLAGS = -buildmode=c-shared
RUSTFLAGS = --release

# Default target: build everything
.PHONY: all
all: lib vm build-samples build-benchmark

# Build the shared library (libzscript.so)
.PHONY: lib
lib:
	@mkdir -p $(BIN_DIR)
	@go build $(GOFLAGS) -o $(BIN_DIR)/$(LIB_NAME) $(SRC_DIR)/lib/corelib.go
	@echo "✅ Library build complete! Library located at $(BIN_DIR)/$(LIB_NAME)"

# Build the main binary (entry-point: cmd/vm/vm.go)
.PHONY: vm
vm:
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(BINARY_NAME) $(SRC_DIR)/vm/vm.go
	@echo "✅ Main build complete! Binary located at $(BIN_DIR)/$(BINARY_NAME)"

# Build the benchmark binary (entry-point: benchmark/benchmark.go)
.PHONY: build-benchmark
build-benchmark:
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/benchmark_zvm benchmark/benchmark.go
	@echo "✅ Benchmark build complete! Binary located at $(BIN_DIR)/benchmark_zvm"

# Build sample programs (C, C++, Go, Rust)
.PHONY: build-samples
build-samples: build-sample-c build-sample-cpp build-sample-go build-sample-rust

# Build C sample
.PHONY: build-sample-c
build-sample-c:
	@mkdir -p $(BIN_DIR)
	@$(CC) $(SAMPLE_DIR)/c/sample.c -o $(BIN_DIR)/sample_c $(CFLAGS)
	@echo "✅ C sample build complete! Binary located at $(BIN_DIR)/sample_c"

# Build C++ sample
.PHONY: build-sample-cpp
build-sample-cpp:
	@mkdir -p $(BIN_DIR)
	@$(CXX) $(SAMPLE_DIR)/cpp/sample.cpp -o $(BIN_DIR)/sample_cpp $(CXXFLAGS)
	@echo "✅ C++ sample build complete! Binary located at $(BIN_DIR)/sample_cpp"

# Build Go sample
.PHONY: build-sample-go
build-sample-go:
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/sample_go $(SAMPLE_DIR)/go/sample.go
	@echo "✅ Go sample build complete! Binary located at $(BIN_DIR)/sample_go"

# Build Rust sample
.PHONY: build-sample-rust
build-sample-rust:
	@mkdir -p $(BIN_DIR)
	@cd $(RUST_DIR) && cargo build $(RUSTFLAGS) && cp target/release/$(RUST_BINARY) ../../../$(BIN_DIR)/
	@echo "✅ Rust sample build complete! Binary located at $(BIN_DIR)/$(RUST_BINARY)"

# Convenience target to build both the shared library and VM executable
.PHONY: build
build: lib vm

# Run the main executable
.PHONY: run
run: vm
	@echo "🚀 Running $(BINARY_NAME)..."
	@$(BIN_DIR)/$(BINARY_NAME)
	@echo "🏁 Execution finished!"

# Run the benchmark executable
.PHONY: run-benchmark
run-benchmark: build-benchmark
	@echo "🚀 Running benchmark_zvm..."
	@$(BIN_DIR)/benchmark_zvm
	@echo "🏁 Execution finished!"

# Run sample programs
.PHONY: run-samples
run-samples: run-sample-c run-sample-cpp run-sample-go run-sample-rust

# Run C sample
.PHONY: run-sample-c
run-sample-c: build-sample-c
	@echo "🚀 Running sample_c..."
	@LD_LIBRARY_PATH=$(BIN_DIR) $(BIN_DIR)/sample_c
	@echo "🏁 Execution finished!"

# Run C++ sample
.PHONY: run-sample-cpp
run-sample-cpp: build-sample-cpp
	@echo "🚀 Running sample_cpp..."
	@LD_LIBRARY_PATH=$(BIN_DIR) $(BIN_DIR)/sample_cpp
	@echo "🏁 Execution finished!"

# Run Go sample
.PHONY: run-sample-go
run-sample-go: build-sample-go
	@echo "🚀 Running sample_go..."
	@LD_LIBRARY_PATH=$(BIN_DIR) $(BIN_DIR)/sample_go
	@echo "🏁 Execution finished!"

# Run Rust sample
.PHONY: run-sample-rust
run-sample-rust: build-sample-rust
	@echo "🚀 Running sample_rust..."
	@LD_LIBRARY_PATH=$(BIN_DIR) $(BIN_DIR)/$(RUST_BINARY)
	@echo "🏁 Execution finished!"

# Test all packages
.PHONY: test
test:
	@echo "🧪 Running tests..."
	@go test -v $(PACKAGE)
	@echo "✅ Tests completed!"

# Build and run the main executable
.PHONY: build-run
build-run: vm
	@echo "🚀 Running $(BINARY_NAME)..."
	@$(BIN_DIR)/$(BINARY_NAME)
	@echo "🏁 Execution finished!"

# Build, test, and run the main executable
.PHONY: build-test-run
build-test-run: vm test
	@echo "🚀 Running $(BINARY_NAME)..."
	@$(BIN_DIR)/$(BINARY_NAME)
	@echo "🏁 Execution finished!"

# Clean up the bin directory
.PHONY: clean
clean:
	@echo "🧹 Cleaning up..."
	@rm -rf $(BIN_DIR)
	@cd $(RUST_DIR) && cargo clean
	@echo "✅ Cleanup complete!"

# Help message
.PHONY: help
help:
	@echo "📜 Makefile targets:"
	@echo "  make             - Build library, main, benchmark, and sample binaries (default)"
	@echo "  make build       - Build the VM and library"
	@echo "  make lib         - Build the shared library into $(BIN_DIR)/$(LIB_NAME)"
	@echo "  make vm          - Build the main binary into $(BIN_DIR)/$(BINARY_NAME)"
	@echo "  make build-benchmark - Build the benchmark binary into $(BIN_DIR)/benchmark_zvm"
	@echo "  make build-samples - Build all sample binaries (C, C++, Go, Rust)"
	@echo "  make build-sample-c - Build the C sample binary into $(BIN_DIR)/sample_c"
	@echo "  make build-sample-cpp - Build the C++ sample binary into $(BIN_DIR)/sample_cpp"
	@echo "  make build-sample-go - Build the Go sample binary into $(BIN_DIR)/sample_go"
	@echo "  make build-sample-rust - Build the Rust sample binary into $(BIN_DIR)/$(RUST_BINARY)"
	@echo "  make run         - Build and run the main executable 🚀"
	@echo "  make run-benchmark - Build and run the benchmark executable 🚀"
	@echo "  make run-samples - Run all sample binaries 🚀"
	@echo "  make run-sample-c - Run the C sample binary 🚀"
	@echo "  make run-sample-cpp - Run the C++ sample binary 🚀"
	@echo "  make run-sample-go - Run the Go sample binary 🚀"
	@echo "  make run-sample-rust - Run the Rust sample binary 🚀"
	@echo "  make test        - Run all tests 🧪"
	@echo "  make build-run   - Build and run the main executable 🚀"
	@echo "  make build-test-run - Build, test, and run the main executable 🚀🧪"
	@echo "  make clean       - Remove the bin directory and clean Rust artifacts 🧹"
	@echo "  make help        - Show this help message 📜"