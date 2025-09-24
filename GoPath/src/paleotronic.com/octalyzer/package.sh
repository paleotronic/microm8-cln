#!/usr/bin/env bash

iconsrc="package/burgericon.png"
icondest="package/icons"
icnspath="package/macos/microM8.app/Contents/Resources"
binpath="package/macos/microM8.app/Contents/MacOS"
iconwin="package/win32/app.ico"
winmanifest="package/win32/versioninfo.json"

GOVERSIONINFO=`which goversioninfo`
CONVERT=`which convert`
PNG2ICNS=`which png2icns`

generate_icons() {
	mkdir "$icondest"
	for size in 256 128 48 32 16; do 
		$CONVERT -resize ${size}x${size} "${iconsrc}" "${icondest}/icon_${size}px.png"
	done
}

generate_icns() {
  	mkdir -p "$icnspath"	
	$PNG2ICNS "$icnspath/microm8.icns" $icondest/*.png
}

generate_ico() {
	$CONVERT "$icondest/icon_256px.png" -define icon:auto-resize="256,32,16"  "$iconwin"	
}

generate_win32_res() {
	$GOVERSIONINFO -64 -o resource64.syso "$winmanifest"
	$GOVERSIONINFO -o resource.syso "$winmanifest"	
}

package_binary() {
	mkdir -p "$binpath"
	cp "$1" "$binpath/microM8"
}

show_help() {
	echo "$0 <mode> <options>"
	echo "  macos <binary>    Package macos binary into app package"
	echo "  windows           Write windows resource.syso"
	exit 0
}

cmd="$1"
if [ "$cmd" == "" ]; then
	cmd="help"
fi

case $cmd in
	windows) 
		#generate_icons
		#generate_ico
		generate_win32_res
		;;
	macos)
		#generate_icons
		#generate_icns
		package_binary $2
		;;
	*)
		show_help
		;;
esac
