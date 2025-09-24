#!/bin/sh

HOST="10.0.0.247"
ACCT="melody"
BASE="Projects"

rsync -av  --exclude=flatpak  ../../paleotronic.com/* "$ACCT@$HOST:$BASE/microm8/GoPath/src/paleotronic.com/"
# ssh "$ACCT@$HOST" "cd ${BASE}/microm8/GoPath/src/paleotronic.com/octalyzer && ./mac-build.sh"
