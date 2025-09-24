#!/bin/bash

# Install Flatpak dependencies for building microM8

RUNTIME_VERSION="23.08"

echo "Installing Flatpak dependencies for microM8..."

# Check if flatpak is installed
if ! command -v flatpak &> /dev/null; then
    echo "Error: Flatpak is not installed."
    echo "Please install Flatpak first:"
    echo "  Ubuntu/Debian: sudo apt install flatpak"
    echo "  Fedora: sudo dnf install flatpak"
    echo "  Arch: sudo pacman -S flatpak"
    exit 1
fi

# Add Flathub repository
echo "Adding Flathub repository..."
flatpak remote-add --if-not-exists --user flathub https://flathub.org/repo/flathub.flatpakrepo

# Install runtime and SDK
echo "Installing Freedesktop Platform runtime..."
flatpak install --user -y flathub org.freedesktop.Platform//$RUNTIME_VERSION

echo "Installing Freedesktop SDK..."
flatpak install --user -y flathub org.freedesktop.Sdk//$RUNTIME_VERSION

echo "Installing Go SDK extension..."
flatpak install --user -y flathub org.freedesktop.Sdk.Extension.golang//$RUNTIME_VERSION

echo ""
echo "All dependencies installed successfully!"
echo "You can now run ./build-flatpak.sh to build microM8"