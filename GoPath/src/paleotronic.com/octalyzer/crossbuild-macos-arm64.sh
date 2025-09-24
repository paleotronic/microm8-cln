#!/bin/sh

HOST="10.0.0.234"
ACCT="melodygriffiths"
BASE="Code"

rsync -av --exclude=flatpak ../../paleotronic.com/* "$ACCT@$HOST:$BASE/microm8/GoPath/src/paleotronic.com/"
#ssh "$ACCT@$HOST" "cd ${BASE}/microm8/GoPath/src/paleotronic.com/octalyzer && ./mac-build.sh"
