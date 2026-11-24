#!/bin/sh
set -eu
HERE=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
umask 077
export XUI_DB_FOLDER="$HERE/state"
export XUI_LOG_FOLDER="$HERE/logs"
export XUI_BIN_FOLDER="$HERE/bin"
export XUI_MAIN_FOLDER="$HERE"
mkdir -p "$XUI_DB_FOLDER" "$XUI_LOG_FOLDER"
"$HERE/xpanel-bootstrap"
printf '%s\n' 'Xpanel local launch. Open http://127.0.0.1:2053. This does not install a system service.'
exec "$HERE/xpanel" run
