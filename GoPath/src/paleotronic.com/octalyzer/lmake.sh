#!/usr/bin/env bash

EXENAME="microM8"
export BASEDIR=`git rev-parse --show-toplevel`
#export GOPATH="${BASEDIR}/GoPath:${BASEDIR}/GoPath/vendor"
if [ "$OS" == "Windows_NT" ]
then
	EXENAME="${EXENAME}.exe"
    export GOPATH="${BASEDIR}\\GoPath;${BASEDIR}\\GoPath\\vendor"
fi

#export GO111MODULE=off

echo $GOPATH

GO=`which go`
XGO=`which xgo`
unset GOROOT

RESOURCEFILES="resource64.syso resource.syso"

buildasm() {
	echo "Starting build for $EXENAME"
	#rm $RESOURCEFILES
	# ./assetbuild.sh
	pushd .
	cd ../core/hardware/cpu/mos6502/asm/cmd
	go build -i -o asm65xx.exe .
	popd
	cp ../core/hardware/cpu/mos6502/asm/cmd/asm65xx.exe .
	#git checkout $RESOURCEFILES
}

build() {
	echo "Starting build for $EXENAME"
	#rm $RESOURCEFILES
	#./assetbuild.sh
	$GO build -o $EXENAME .
	#git checkout $RESOURCEFILES
}

buildmacx86() {
	echo "Starting build for $EXENAME"
	#rm $RESOURCEFILES
	./assetbuild.sh
# 	cd ../../../..
	GOOS=darwin GOARCH=amd64 $XGO build -o $EXENAME .
	#git checkout $RESOURCEFILES
}

remint() {
	EXENAME="remint"
	echo "Starting build for $EXENAME"
	#rm $RESOURCEFILES
	./assetbuild.sh
	go build -i -tags remint -o $EXENAME .
	#git checkout $RESOURCEFILES
}

build_nox() {
        EXENAME="noxarchaist"
		if [ "$OS" == "Windows_NT" ]
		then
			EXENAME="${EXENAME}.exe"
		fi
        echo "Starting build for $EXENAME"
        #rm $RESOURCEFILES
        ./assetbuild.sh
        go build -i -tags nox -o $EXENAME .
        #git checkout $RESOURCEFILES
}

profile() {
	build
	./${EXENAME} -inst-vms
}

run() {
	build 
	./${EXENAME}
}

case $1 in
	profile)
		profile
		;;
	remint) 
		remint
		;;
	asm)
		buildasm
		;;
	macos)
		buildmacx86
		;;
	nox)
		build_nox
		;;
	run) 
		run
		;;
	*)
		build
		;;
esac

