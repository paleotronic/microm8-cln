#!/bin/bash

# Package microM8 for macOS
# This script builds microM8 and packages it as a macOS .app bundle

set -e

echo "Generating version information..."
./gen-version.sh

echo "Building microM8..."
./lmake.sh build

if [ ! -f "./microM8" ]; then
    echo "Error: microM8 binary not found after build"
    exit 1
fi

echo "Preparing app bundle..."

# Create the app bundle directory structure
APP_DIR="package/macos/microM8.app"
CONTENTS_DIR="$APP_DIR/Contents"
MACOS_DIR="$CONTENTS_DIR/MacOS"

# Create MacOS directory if it doesn't exist
mkdir -p "$MACOS_DIR"

# Copy the binary to the app bundle
cp ./microM8 "$MACOS_DIR/microM8"

# Make sure the binary is executable
chmod +x "$MACOS_DIR/microM8"

# Copy RELEASE.md to the package directory
if [ -f "./RELEASE.md" ]; then
    cp ./RELEASE.md "package/macos/RELEASE.md"
fi

# Create the zip file
ZIP_NAME="microM8-macos-$(go env GOARCH)-$(date +%Y%m%d).zip"
echo "Creating $ZIP_NAME..."

# Change to package/macos directory to create a clean zip structure
cd package/macos
if [ -f "RELEASE.md" ]; then
    zip -r "../../$ZIP_NAME" microM8.app RELEASE.md
else
    zip -r "../../$ZIP_NAME" microM8.app
fi
cd ../..

echo "Successfully created $ZIP_NAME"
echo "App bundle is ready at: $APP_DIR"
