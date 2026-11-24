#!/bin/sh
set -eu
HERE=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd -P)
ROOT=$(CDPATH= cd -- "$HERE/../.." && pwd -P)
RELEASE=$(basename -- "$HERE")
umask 077
[ -f "$ROOT/INSTALL.json" ] && [ -d "$ROOT/runtime-bin/$RELEASE" ] || { printf '%s\n' 'Use the owned local installer before this managed launcher'; exit 1; }
command -v flock >/dev/null 2>&1 || { printf '%s\n' 'flock is required for managed starts'; exit 1; }
exec 9>"$ROOT/state/.xpanel-runtime.lock"
flock -n 9 || { printf '%s\n' 'Another Xpanel process or update is using this state'; exit 1; }
export XUI_DB_FOLDER="$ROOT/state"
export XUI_LOG_FOLDER="$ROOT/logs"
export XUI_BIN_FOLDER="$ROOT/runtime-bin/$RELEASE"
export XUI_MAIN_FOLDER="$HERE"
"$HERE/xpanel-bootstrap"
printf '%s\n' 'Xpanel managed local launch. Open http://127.0.0.1:2053.'
exec "$HERE/xpanel" run
