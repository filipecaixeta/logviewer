#!/bin/bash

# Detect OS and Architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

echo "Detected OS: $OS, Architecture: $ARCH"

# Define the download URL base
BASE_URL="https://github.com/filipecaixeta/logviewer/releases/latest/download"

# Determine the correct file name based on OS and Architecture
FILENAME=""

case "$OS" in
    Linux*)     FILENAME="logviewer-linux-$ARCH" ;;
    Darwin*)    FILENAME="logviewer-macos-$ARCH" ;;
    *)          echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Construct the full download URL
DOWNLOAD_URL="${BASE_URL}/${FILENAME}"

echo "Download URL: $DOWNLOAD_URL"

# Download the file
curl -L $DOWNLOAD_URL -o logviewer

# Add execute permissions
chmod +x logviewer

# Move the file to a bin directory
sudo mv logviewer /usr/local/bin/

# macOS-specific steps to authorize the binary
if [ "$OS" == "Darwin" ]; then
    echo "Authorizing the binary on macOS..."

    # Use xattr to remove the "com.apple.quarantine" attribute
    sudo xattr -dr com.apple.quarantine /usr/local/bin/logviewer

    echo "Binary authorized on macOS. You may still need to allow it in System Preferences > Security & Privacy."
fi

echo "Installation completed. Run with 'logviewer'"
