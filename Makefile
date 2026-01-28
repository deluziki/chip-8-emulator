# CHIP-8 Emulator Makefile

BINARY_NAME=chip8-emulator
GO=go

.PHONY: all build clean run deps

all: build

# Build the emulator
build:
	$(GO) build -o $(BINARY_NAME) .

# Build with race detector (for development)
build-race:
	$(GO) build -race -o $(BINARY_NAME) .

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	$(GO) clean

# Install dependencies
deps:
	$(GO) mod download
	$(GO) mod tidy

# Run the emulator (requires ROM argument)
run: build
	@if [ -z "$(ROM)" ]; then \
		echo "Usage: make run ROM=<path-to-rom>"; \
		echo "Example: make run ROM=roms/PONG"; \
	else \
		./$(BINARY_NAME) $(ROM); \
	fi

# Run with custom scale
run-scaled: build
	@if [ -z "$(ROM)" ]; then \
		echo "Usage: make run-scaled ROM=<path-to-rom> SCALE=<scale>"; \
	else \
		./$(BINARY_NAME) -scale $(or $(SCALE),15) $(ROM); \
	fi

# Run tests
test:
	$(GO) test -v ./...

# Format code
fmt:
	$(GO) fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Show help
help:
	@echo "CHIP-8 Emulator - Build Targets"
	@echo ""
	@echo "  make build     - Build the emulator"
	@echo "  make clean     - Remove build artifacts"
	@echo "  make deps      - Download and tidy dependencies"
	@echo "  make run ROM=<path>  - Build and run with specified ROM"
	@echo "  make test      - Run tests"
	@echo "  make fmt       - Format source code"
	@echo "  make help      - Show this help message"
