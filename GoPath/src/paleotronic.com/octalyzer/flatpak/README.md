# microM8 Flatpak

This directory contains the files needed to build microM8 as a Flatpak application.

## Prerequisites

To build the Flatpak, you need:

1. **Install Flatpak and Flatpak Builder**:
   ```bash
   # Ubuntu/Debian
   sudo apt install flatpak flatpak-builder
   
   # Fedora
   sudo dnf install flatpak flatpak-builder
   
   # Arch
   sudo pacman -S flatpak flatpak-builder
   ```

2. **Install required runtimes** (automatic or manual):
   
   **Option A - Automatic** (recommended):
   ```bash
   cd flatpak
   ./install-dependencies.sh
   ```
   
   **Option B - Manual**:
   ```bash
   # Add Flathub repository
   flatpak remote-add --if-not-exists flathub https://flathub.org/repo/flathub.flatpakrepo
   
   # Install required runtime and SDK
   flatpak install flathub org.freedesktop.Platform//23.08 org.freedesktop.Sdk//23.08
   flatpak install flathub org.freedesktop.Sdk.Extension.golang//23.08
   ```

## Building

To build the Flatpak:

```bash
cd flatpak
./build-flatpak.sh
```

This will:
1. Build microM8 from source
2. Package it as a Flatpak
3. Create a `.flatpak` bundle file for distribution
4. Optionally install it locally

## Installing

To install the built Flatpak:

```bash
flatpak install --user com.paleotronic.microM8.flatpak
```

## Running

To run microM8:

```bash
flatpak run com.paleotronic.microM8
```

## Files

- `com.paleotronic.microM8.yml` - Flatpak manifest file
- `com.paleotronic.microM8.appdata.xml` - AppStream metadata
- `com.paleotronic.microM8.desktop` - Desktop entry file
- `build-flatpak.sh` - Build script with dependency checking
- `install-dependencies.sh` - Script to install required runtimes
- `README.md` - This file

## Dependencies

The Flatpak build includes:
- PortAudio library for audio support
- Go SDK extension for building from source

## Permissions

The Flatpak requests the following permissions:
- X11 and Wayland display access
- Audio (PulseAudio)
- OpenGL/DRI for graphics
- Network access (for MCP features)
- Home directory access (for loading disk images)

## Distribution

The generated `.flatpak` bundle file can be distributed to users who can install it with:

```bash
flatpak install --user microM8.flatpak
```