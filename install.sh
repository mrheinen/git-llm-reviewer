#!/bin/bash
set -e

# Installation script for git-llm-review

# Default values
INSTALL_DIR="/usr/local/bin"
VERSION="latest"
TEMP_DIR=$(mktemp -d)
CLEANUP=true
BINARY_NAME="git-llm-review"

# ANSI color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print messages
print_message() {
    echo -e "${BLUE}[GIT-LLM-REVIEW]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[GIT-LLM-REVIEW]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[GIT-LLM-REVIEW]${NC} $1"
}

print_error() {
    echo -e "${RED}[GIT-LLM-REVIEW]${NC} $1"
}

# Function to cleanup temporary files
cleanup() {
    if [ "$CLEANUP" = true ]; then
        print_message "Cleaning up temporary files..."
        rm -rf "$TEMP_DIR"
    fi
}

# Set up trap to cleanup on exit
trap cleanup EXIT

# Function to detect platform
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    # Convert architectures to standard naming
    case "$ARCH" in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    # Verify OS is supported
    case "$OS" in
        linux|darwin)
            ;;
        *)
            print_error "Unsupported operating system: $OS"
            exit 1
            ;;
    esac

    PLATFORM="${OS}-${ARCH}"
    print_message "Detected platform: $PLATFORM"
}

# Function to download release
download_release() {
    local release_url="https://github.com/niels/git-llm-review/releases/download/${VERSION}/git-llm-review-${VERSION}-${PLATFORM}.zip"
    
    print_message "Downloading git-llm-review ${VERSION} for ${PLATFORM}..."
    curl -L -o "${TEMP_DIR}/release.zip" "$release_url" || {
        print_error "Failed to download release. Please check if the version and platform are correct."
        exit 1
    }
    
    print_message "Extracting release..."
    unzip -q "${TEMP_DIR}/release.zip" -d "$TEMP_DIR" || {
        print_error "Failed to extract release."
        exit 1
    }
}

# Function to install binary
install_binary() {
    print_message "Installing git-llm-review to $INSTALL_DIR..."
    
    # Find the binary in the extracted files
    BINARY_PATH=$(find "$TEMP_DIR" -name "$BINARY_NAME*" -type f -executable | head -n 1)
    
    if [ -z "$BINARY_PATH" ]; then
        print_error "Binary not found in the downloaded package."
        exit 1
    fi
    
    # Check if installation directory exists and is writable
    if [ ! -d "$INSTALL_DIR" ]; then
        print_warning "$INSTALL_DIR does not exist. Creating directory..."
        mkdir -p "$INSTALL_DIR" || {
            print_error "Failed to create directory. Try running with sudo."
            exit 1
        }
    fi
    
    # Install binary
    cp "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME" || {
        print_error "Failed to install binary. Try running with sudo."
        exit 1
    }
    
    chmod +x "$INSTALL_DIR/$BINARY_NAME" || {
        print_error "Failed to make binary executable. Try running with sudo."
        exit 1
    }
    
    print_success "git-llm-review has been installed to $INSTALL_DIR/$BINARY_NAME"
}

# Function to verify installation
verify_installation() {
    if command -v "$INSTALL_DIR/$BINARY_NAME" >/dev/null 2>&1; then
        print_success "Verification successful! You can now use git-llm-review."
        "$INSTALL_DIR/$BINARY_NAME" --version
    else
        print_warning "Binary installed but not found in PATH. You may need to add $INSTALL_DIR to your PATH."
    fi
}

# Parse command-line arguments
while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in
        --dir)
            INSTALL_DIR="$2"
            shift
            shift
            ;;
        --version)
            VERSION="$2"
            shift
            shift
            ;;
        --no-cleanup)
            CLEANUP=false
            shift
            ;;
        --help)
            echo "Usage: ./install.sh [OPTIONS]"
            echo "OPTIONS:"
            echo "  --dir DIR        Set the installation directory (default: /usr/local/bin)"
            echo "  --version VER    Set the version to install (default: latest)"
            echo "  --no-cleanup     Don't remove temporary files after installation"
            echo "  --help           Show this help message"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Main installation process
print_message "Starting installation of git-llm-review..."
detect_platform
download_release
install_binary
verify_installation

print_success "Installation completed successfully!"
echo ""
print_message "To get started, run: git-llm-review --help"
echo ""
