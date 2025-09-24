#!/bin/sh

source $HOME/.zshrc

# generate version info
./gen-version.sh

# remove syso
rm *.syso

# build
./lmake.sh build

# package
./package_macos.sh
