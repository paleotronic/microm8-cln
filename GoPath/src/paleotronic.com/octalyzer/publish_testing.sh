#!/bin/sh

dist="testing"
host="build@update.paleotronic.com"
version=`grep PERCOL8_BUILD ../update/build_id.go | awk '{print $3}' | sed 's/"//g'`
platform="$1"
arch="$2"
target="$platform-$arch"
mkdir -p "next"
exename="microM8"

sha="sha256sum"

if [ "$platform" = "windows" ]; then
    exename="microM8.exe"
fi

if [ "$platform" = "darwin" ]; then
    exename="microM8"
    sha="shasum -a 256"
fi

if [ "$platform" = "linux" ]; then
    strip -s $exename
fi

$sha $exename > "next/checksum.sha256"

xz -k $exename

mv $exename.xz "next/package.xz"
echo $version > "next/version"
ls -l "next/"

base="/home/build/dist/${dist}/${platform}/${arch}"
existing="${base}/current"
next="${base}/next"

scp -r ./next "${host}:${base}/"
ssh ${host} "cd ${base} && prev=\`cat ./current/version\` && mv ./current \$prev && mv next current"

