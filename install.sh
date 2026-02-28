#!/bin/bash
set -euo pipefail

REPO="hiragram/agent-workspace"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
COMMAND_NAME="aw"

echo "Installing $COMMAND_NAME..."

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux|darwin) ;;
  *)
    echo "Error: Unsupported OS: $OS" >&2
    exit 1
    ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "Error: Unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

# Get latest release tag
LATEST=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$LATEST" ]; then
  echo "Error: Could not determine latest release" >&2
  exit 1
fi

URL="https://github.com/$REPO/releases/download/$LATEST/${COMMAND_NAME}_${OS}_${ARCH}.tar.gz"

echo "Downloading $COMMAND_NAME $LATEST for ${OS}/${ARCH}..."

# Download and extract
mkdir -p "$INSTALL_DIR"
curl -fsSL "$URL" | tar xz -C "$INSTALL_DIR" "$COMMAND_NAME"
chmod +x "$INSTALL_DIR/$COMMAND_NAME"

# PATH check
if ! echo "$PATH" | tr ':' '\n' | grep -qx "$INSTALL_DIR"; then
  echo ""
  echo "Warning: $INSTALL_DIR is not in your PATH."
  echo "Add the following to your shell profile (~/.zshrc, ~/.bashrc, etc.):"
  echo ""
  echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
  echo ""
fi

echo "Installed $COMMAND_NAME $LATEST to $INSTALL_DIR/$COMMAND_NAME"
echo ""
echo "Run '$COMMAND_NAME' to start agent-workspace."
