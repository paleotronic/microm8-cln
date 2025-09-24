#!/bin/bash

# Package microM8 for Linux
# This script builds microM8 and packages it as a zip file

set -e

echo "Generating version information..."
./gen-version.sh

echo "Building microM8 for Linux..."
./lmake.sh build

if [ ! -f "./microM8" ]; then
    echo "Error: microM8 binary not found after build"
    exit 1
fi

echo "Preparing Linux package..."

# Create a temporary directory for packaging
PACKAGE_DIR="package/linux/microM8"
mkdir -p "$PACKAGE_DIR"

# Copy the binary
cp ./microM8 "$PACKAGE_DIR/microM8"

# Make sure the binary is executable
chmod +x "$PACKAGE_DIR/microM8"

# Copy RELEASE.md if it exists
if [ -f "./RELEASE.md" ]; then
    cp ./RELEASE.md "$PACKAGE_DIR/RELEASE.md"
fi

# Copy desktop file if it exists
if [ -f "./microM8.desktop" ]; then
    cp ./microM8.desktop "$PACKAGE_DIR/microM8.desktop"
fi

# Copy icon file if it exists
if [ -f "./microM8.png" ]; then
    cp ./microM8.png "$PACKAGE_DIR/microM8.png"
fi

# Create the zip file with the requested naming convention
ZIP_NAME="microM8-linux-$(go env GOARCH)-$(date +%Y%m%d).zip"
echo "Creating $ZIP_NAME..."

# Change to package/linux directory to create a clean zip structure
cd package/linux
zip -r "../../$ZIP_NAME" microM8
cd ../..

# Clean up temporary directory
rm -rf "$PACKAGE_DIR"

echo "Successfully created $ZIP_NAME"
