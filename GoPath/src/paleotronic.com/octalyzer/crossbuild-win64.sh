#!/bin/bash

export BASEDIR=`git rev-parse --show-toplevel`
export GOPATH="$BASEDIR/GoPath:$BASEDIR/GoPath/vendor"
export GOOS="windows"
export GOARCH="amd64"

export GOBUILDFLAGS="-x -ldflags '-extldflags \"-static\"'"

export MXEMODE="static"
export CCWINCROSS="/usr/bin/x86_64-w64-mingw32-gcc"

export CGO_ENABLED=1
export CC=$CCWINCROSS
export GOOS=windows
export GOARCH=amd64
export EXENAME=microM8.exe
export LD_LIBRARY_PATH="/usr/x86_64-w64-mingw32/lib"
export PKG_CONFIG_PATH="/usr/x86_64-w64-mingw32/lib/pkgconfig"

cat <<EOF
==============================================================================
CROSS BUILDING MICROM8 FOR WIN64/X86
------------------------------------------------------------------------------
Repo Base: $BASEDIR
Output   : $EXENAME
Run with : wine ./$EXENAME
==============================================================================
EOF

cp resources/microm8/amd64/resource.syso .
GO111MODULE=off go build -ldflags '-extldflags "-static"' -o "${EXENAME}" .
