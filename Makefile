# Build settings
BINARY_NAME=git-llm-review
OUTPUT_DIR=dist
PACKAGE_DIR=package

# Version information
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.0-dev")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH=$(shell git symbolic-ref --short -q HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go build flags
GO=go
LDFLAGS=-ldflags "-X github.com/niels/git-llm-review/pkg/version.Version=${VERSION} \
                  -X github.com/niels/git-llm-review/pkg/version.GitCommit=${GIT_COMMIT} \
                  -X github.com/niels/git-llm-review/pkg/version.BuildDate=${BUILD_DATE}"
BUILD_FLAGS=-trimpath ${LDFLAGS}

# Platforms to compile for
PLATFORMS=linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64

# Default target
.PHONY: all
all: clean test build

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -rf ${OUTPUT_DIR}
	@rm -rf ${PACKAGE_DIR}

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@${GO} test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@${GO} test -v -coverprofile=coverage.out ./...
	@${GO} tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build for the current platform
.PHONY: build
build:
	@echo "Building ${BINARY_NAME} version ${VERSION} (${GIT_COMMIT})..."
	@mkdir -p ${OUTPUT_DIR}
	@${GO} build ${BUILD_FLAGS} -o ${OUTPUT_DIR}/${BINARY_NAME} ./cmd/git-llm-review

# Build for all platforms
.PHONY: build-all
build-all: $(addprefix build-,${PLATFORMS})

# Build for a specific platform
.PHONY: $(addprefix build-,${PLATFORMS})
$(addprefix build-,${PLATFORMS}):
	$(eval PLATFORM := $(subst build-,,$@))
	$(eval OS := $(firstword $(subst -, ,${PLATFORM})))
	$(eval ARCH := $(lastword $(subst -, ,${PLATFORM})))
	$(eval BINARY_SUFFIX := $(if $(filter windows,${OS}),.exe,))
	@echo "Building ${BINARY_NAME} for ${OS}/${ARCH}..."
	@mkdir -p ${OUTPUT_DIR}/${OS}-${ARCH}
	@GOOS=${OS} GOARCH=${ARCH} ${GO} build ${BUILD_FLAGS} \
		-o ${OUTPUT_DIR}/${OS}-${ARCH}/${BINARY_NAME}${BINARY_SUFFIX} \
		./cmd/git-llm-review

# Create release packages
.PHONY: release
release: clean build-all package

# Create packages for distribution
.PHONY: package
package: $(addprefix package-,${PLATFORMS})

# Package for a specific platform
.PHONY: $(addprefix package-,${PLATFORMS})
$(addprefix package-,${PLATFORMS}):
	$(eval PLATFORM := $(subst package-,,$@))
	$(eval OS := $(firstword $(subst -, ,${PLATFORM})))
	$(eval ARCH := $(lastword $(subst -, ,${PLATFORM})))
	$(eval BINARY_SUFFIX := $(if $(filter windows,${OS}),.exe,))
	@echo "Packaging ${BINARY_NAME} for ${OS}/${ARCH}..."
	@mkdir -p ${PACKAGE_DIR}
	@mkdir -p ${PACKAGE_DIR}/${BINARY_NAME}-${VERSION}-${OS}-${ARCH}
	@cp ${OUTPUT_DIR}/${OS}-${ARCH}/${BINARY_NAME}${BINARY_SUFFIX} ${PACKAGE_DIR}/${BINARY_NAME}-${VERSION}-${OS}-${ARCH}/
	@cp README.md ${PACKAGE_DIR}/${BINARY_NAME}-${VERSION}-${OS}-${ARCH}/
	@cp LICENSE ${PACKAGE_DIR}/${BINARY_NAME}-${VERSION}-${OS}-${ARCH}/ 2>/dev/null || echo "No LICENSE file found"
	@cp -r docs ${PACKAGE_DIR}/${BINARY_NAME}-${VERSION}-${OS}-${ARCH}/ 2>/dev/null || echo "No docs directory found"
	@cd ${PACKAGE_DIR} && zip -r ${BINARY_NAME}-${VERSION}-${OS}-${ARCH}.zip ${BINARY_NAME}-${VERSION}-${OS}-${ARCH}
	@echo "Package created: ${PACKAGE_DIR}/${BINARY_NAME}-${VERSION}-${OS}-${ARCH}.zip"

# Install locally
.PHONY: install
install: build
	@echo "Installing ${BINARY_NAME}..."
	@cp ${OUTPUT_DIR}/${BINARY_NAME} $(GOPATH)/bin/${BINARY_NAME} 2>/dev/null || \
		cp ${OUTPUT_DIR}/${BINARY_NAME} /usr/local/bin/${BINARY_NAME} 2>/dev/null || \
		(echo "ERROR: Could not install binary. Please manually copy ${OUTPUT_DIR}/${BINARY_NAME} to your PATH.")

# Generate documentation
.PHONY: docs
docs:
	@echo "Generating documentation..."
	@mkdir -p docs
	@echo "# ${BINARY_NAME} User Guide\n\nVersion: ${VERSION}\n" > docs/user_guide.md
	@echo "## Installation\n\n\`\`\`\ncp ${BINARY_NAME} /usr/local/bin/\n\`\`\`\n" >> docs/user_guide.md
	@echo "## Usage\n\n\`\`\`\n${BINARY_NAME} --help\n\`\`\`\n" >> docs/user_guide.md
	@echo "Documentation generated in docs directory"

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all            - Clean, run tests, and build for current platform"
	@echo "  clean          - Remove build artifacts"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  build          - Build for current platform"
	@echo "  build-all      - Build for all platforms"
	@echo "  release        - Create release packages for all platforms"
	@echo "  install        - Install binary locally"
	@echo "  docs           - Generate documentation"
	@echo "  help           - Show this help message"
