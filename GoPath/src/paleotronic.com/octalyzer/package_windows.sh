#!/bin/bash

# Package microM8 for Windows
# This script builds microM8 for Windows and packages it as a zip file
# Can be run from Linux with cross-compilation support

set -e

echo "Generating version information..."
./gen-version.sh

echo "Building microM8 for Windows..."
# Set GOOS and GOARCH for Windows cross-compilation
export GOOS=windows
export GOARCH=amd64
./crossbuild-win64.sh

# Check for Windows executable
if [ ! -f "./microM8.exe" ]; then
    echo "Error: microM8.exe binary not found after build"
    exit 1
fi

echo "Preparing Windows package..."

# Create a temporary directory for packaging
PACKAGE_DIR="package/windows/microM8"
mkdir -p "$PACKAGE_DIR"

# Copy the binary (handle both .exe and non-.exe cases)
if [ -f "./microM8.exe" ]; then
    cp ./microM8.exe "$PACKAGE_DIR/microM8.exe"
fi

# Copy RELEASE.md if it exists
if [ -f "./RELEASE.md" ]; then
    cp ./RELEASE.md "$PACKAGE_DIR/RELEASE.md"
fi

# Copy any Windows-specific DLLs if needed
# (Add any required DLL copies here if necessary)

# Create the zip file with the requested naming convention
ZIP_NAME="microM8-windows-$(go env GOARCH)-$(date +%Y%m%d).zip"
echo "Creating $ZIP_NAME..."

# Change to package/windows directory to create a clean zip structure
cd package/windows
zip -r "../../$ZIP_NAME" microM8
cd ../..

# Clean up temporary directory
rm -rf "$PACKAGE_DIR"

# Reset GOOS and GOARCH
unset GOOS
unset GOARCH

echo "Successfully created $ZIP_NAME"
