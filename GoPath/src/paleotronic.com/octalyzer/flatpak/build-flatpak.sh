#!/bin/bash

# Build script for microM8 Flatpak
# This script builds and optionally installs the microM8 Flatpak

set -e

FLATPAK_ID="com.paleotronic.microM8"
MANIFEST="com.paleotronic.microM8.yml"
REPO_DIR="repo"
BUILD_DIR="build"
RUNTIME_VERSION="23.08"

echo "Building microM8 Flatpak..."

# Check if we're in the flatpak directory
if [ ! -f "$MANIFEST" ]; then
    echo "Error: Must run this script from the flatpak directory"
    echo "Current directory: $(pwd)"
    exit 1
fi

# Function to check if a Flatpak runtime is installed
check_runtime() {
    flatpak list --runtime | grep -q "$1" || return 1
}

# Check for required runtimes and SDK
echo "Checking for required Flatpak runtimes..."

MISSING_DEPS=false

if ! check_runtime "org.freedesktop.Platform/x86_64/$RUNTIME_VERSION"; then
    echo "Missing: org.freedesktop.Platform/x86_64/$RUNTIME_VERSION"
    MISSING_DEPS=true
fi

if ! check_runtime "org.freedesktop.Sdk/x86_64/$RUNTIME_VERSION"; then
    echo "Missing: org.freedesktop.Sdk/x86_64/$RUNTIME_VERSION"
    MISSING_DEPS=true
fi

if ! check_runtime "org.freedesktop.Sdk.Extension.golang/x86_64/$RUNTIME_VERSION"; then
    echo "Missing: org.freedesktop.Sdk.Extension.golang/x86_64/$RUNTIME_VERSION"
    MISSING_DEPS=true
fi

if [ "$MISSING_DEPS" = true ]; then
    echo ""
    echo "Required Flatpak runtimes are not installed."
    echo ""
    echo "To install them, run:"
    echo ""
    echo "  # Add Flathub repository if not already added"
    echo "  flatpak remote-add --if-not-exists flathub https://flathub.org/repo/flathub.flatpakrepo"
    echo ""
    echo "  # Install required runtimes"
    echo "  flatpak install flathub org.freedesktop.Platform//$RUNTIME_VERSION org.freedesktop.Sdk//$RUNTIME_VERSION"
    echo "  flatpak install flathub org.freedesktop.Sdk.Extension.golang//$RUNTIME_VERSION"
    echo ""
    
    read -p "Would you like to install these dependencies now? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "Adding Flathub repository..."
        flatpak remote-add --if-not-exists flathub https://flathub.org/repo/flathub.flatpakrepo
        
        echo "Installing required runtimes..."
        flatpak install -y flathub org.freedesktop.Platform//$RUNTIME_VERSION org.freedesktop.Sdk//$RUNTIME_VERSION
        flatpak install -y flathub org.freedesktop.Sdk.Extension.golang//$RUNTIME_VERSION
    else
        echo "Cannot proceed without required runtimes. Exiting."
        exit 1
    fi
fi

# Clean previous builds
echo "Cleaning previous builds..."
rm -rf "$BUILD_DIR" "$REPO_DIR"

# Initialize repo if it doesn't exist
if [ ! -d "$REPO_DIR" ]; then
    echo "Initializing Flatpak repository..."
    ostree init --mode=archive-z2 --repo="$REPO_DIR"
fi

# Build the Flatpak
echo "Building Flatpak..."
flatpak-builder --force-clean --repo="$REPO_DIR" "$BUILD_DIR" "$MANIFEST"

if [ $? -eq 0 ]; then
    echo "Build successful!"
    
    # Create bundle file for distribution
    BUNDLE_FILE="${FLATPAK_ID}.flatpak"
    echo "Creating bundle file: $BUNDLE_FILE"
    flatpak build-bundle "$REPO_DIR" "$BUNDLE_FILE" "$FLATPAK_ID"
    
    echo ""
    echo "Flatpak built successfully!"
    echo "Bundle file created: $BUNDLE_FILE"
    echo ""
    echo "To install locally, run:"
    echo "  flatpak install --user $BUNDLE_FILE"
    echo ""
    echo "To run the installed application:"
    echo "  flatpak run $FLATPAK_ID"
    
    # Ask if user wants to install
    read -p "Do you want to install the Flatpak now? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "Installing Flatpak..."
        flatpak install --user --reinstall "$BUNDLE_FILE"
        echo "Installation complete!"
    fi
else
    echo "Build failed!"
    exit 1
fi