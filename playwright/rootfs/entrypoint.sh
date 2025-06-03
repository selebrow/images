#!/bin/bash
set -eu

if [ "${ENABLE_VNC:-false}" == true ]; then
    SCREEN_RESOLUTION=${SCREEN_RESOLUTION:-1920x1080x24}
    DISPLAY_NUM=99
    export DISPLAY=:${DISPLAY_NUM}

    /usr/bin/xvfb-run -l -n ${DISPLAY_NUM} -s "-ac -screen 0 ${SCREEN_RESOLUTION} -noreset -listen tcp" /usr/bin/fluxbox -display ${DISPLAY} >/dev/null 2>&1 &

    until wmctrl -m >/dev/null 2>&1; do
      echo Waiting X server...
      sleep 0.1
    done

    x11vnc -display ${DISPLAY} -passwd "${VNC_PASSWORD:-selebrow}" -shared -forever -loop500 -rfbport 5900 -rfbportv6 5900 >/dev/null 2>&1 &
fi

exec node node_modules/.bin/playwright run-server --port 4444 --host 0.0.0.0
