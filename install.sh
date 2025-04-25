#!/bin/bash
# Script to install or uninstall the ZScript.
# It downloads the latest release binary from GitHub using curl and installs it either:
#   - System-wide in /usr/local/bin (with sudo), or
#   - In $HOME/.local/bin (without sudo) and ensures $HOME/.local/bin is in PATH.

# Variables for GitHub repo
OWNER="cryptrunner49"
REPO="zscript"
FILE="zvm"
DOWNLOAD_URL="https://github.com/$OWNER/$REPO/releases/latest/download/$FILE"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Usage message
usage() {
    echo "Usage: $0 {install|uninstall} [--system|--user]"
    echo "  install   - Install ZScript"
    echo "  uninstall - Uninstall ZScript"
    echo "  --system  - Install system-wide (uses sudo, installs to /usr/local/bin)"
    echo "  --user    - Install for the current user (default; installs to \$HOME/.local/bin)"
    exit 1
}

# Check for at least one argument
if [ "$#" -lt 1 ]; then
    usage
fi

ACTION="$1"
MODE="$2"  # Can be --system or --user; defaults to --user if not provided

# Default to user installation if mode is not specified
if [ -z "$MODE" ]; then
    MODE="--user"
fi

if [ "$MODE" != "--system" ] && [ "$MODE" != "--user" ]; then
    usage
fi

# Set installation directory and whether to use sudo based on mode
if [ "$MODE" == "--system" ]; then
    INSTALL_DIR="/usr/local/bin"
    USE_SUDO=1
else
    INSTALL_DIR="$HOME/.local/bin"
    USE_SUDO=0
fi

# Function to check if a command succeeded
check_status() {
    if [ $? -ne 0 ]; then
        echo -e "${RED}Error: $1 failed${NC}"
        exit 1
    fi
}

# Function to add $HOME/.local/bin to PATH in the default shell RC file if it's not already in PATH
add_local_bin_to_path() {
    if ! echo "$PATH" | grep -q "$HOME/.local/bin"; then
        # Detect default shell and RC file
        if [ -n "$BASH_VERSION" ]; then
            RC_FILE="$HOME/.bashrc"
        elif [ -n "$ZSH_VERSION" ]; then
            RC_FILE="$HOME/.zshrc"
        else
            echo -e "${RED}Warning: Could not detect your shell. Please add \$HOME/.local/bin to your PATH manually.${NC}"
            return
        fi
        echo "export PATH=\$PATH:\$HOME/.local/bin" >> "$RC_FILE"
        echo -e "${GREEN}Added \$HOME/.local/bin to PATH in $RC_FILE.${NC}"
    fi
}

# Install function
install_zscript() {
    echo "Starting ZScript installation..."

    # Create installation directory if it doesn't exist
    if [ "$USE_SUDO" -eq 1 ]; then
        sudo mkdir -p "$INSTALL_DIR" || check_status "Creating system install directory $INSTALL_DIR"
    else
        mkdir -p "$INSTALL_DIR" || check_status "Creating user install directory $INSTALL_DIR"
    fi

    # Download the latest release binary
    echo "Downloading ZScript from: $DOWNLOAD_URL"
    curl -sL -o "$FILE" "$DOWNLOAD_URL" || check_status "Downloading ZScript"

    # Make the binary executable
    chmod +x "$FILE" || check_status "Setting executable permission on $FILE"

    # Move the binary to the installation directory
    if [ "$USE_SUDO" -eq 1 ]; then
        sudo mv "$FILE" "$INSTALL_DIR/zscript" || check_status "Moving binary to $INSTALL_DIR"
    else
        mv "$FILE" "$INSTALL_DIR/zscript" || check_status "Moving binary to $INSTALL_DIR"
        add_local_bin_to_path
    fi

    echo -e "${GREEN}ZScript installed successfully in $INSTALL_DIR!${NC}"
    echo "Run 'zscript --help' for usage instructions."
}

# Uninstall function
uninstall_zscript() {
    echo "Starting ZScript uninstallation..."
    if [ "$USE_SUDO" -eq 1 ]; then
        sudo rm -f "$INSTALL_DIR/zscript" || check_status "Removing binary from $INSTALL_DIR"
    else
        rm -f "$INSTALL_DIR/zscript" || check_status "Removing binary from $INSTALL_DIR"
    fi
    echo -e "${GREEN}ZScript uninstalled successfully from $INSTALL_DIR!${NC}"
}

# Main script logic
case "$ACTION" in
    install)
        install_zscript
        ;;
    uninstall)
        uninstall_zscript
        ;;
    *)
        usage
        ;;
esac

exit 0