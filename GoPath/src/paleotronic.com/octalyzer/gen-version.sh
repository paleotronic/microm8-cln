#!/bin/bash


build=`date -u '+%Y%m%d%H%M'`
revision=`git rev-parse HEAD`
date=`date -u '+%Y-%m-%d_%I:%M:%S%p'`

if [ -f build_id.go ]; then
	rm -f build_id.go
fi

cat <<END > ../update/build_id.go
package update

const (
	PERCOL8_BUILD   = "$build"
	PERCOL8_GITHASH = "$revision"
	PERCOL8_DATE    = "$date"
)

END

