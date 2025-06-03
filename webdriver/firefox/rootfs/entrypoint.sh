#!/bin/bash
set -eu

SCREEN_RESOLUTION=${SCREEN_RESOLUTION:-"1920x1080x24"}
DISPLAY_NUM=99
export DISPLAY=":$DISPLAY_NUM"

VERBOSE=${VERBOSE:-""}
DRIVER_ARGS=${DRIVER_ARGS:-""}
if [ -n "$VERBOSE" ]; then
    DRIVER_ARGS="$DRIVER_ARGS --log debug"
fi

clean() {
  if [ -n "${FILESERVER_PID:-}" ]; then
    kill -TERM "$FILESERVER_PID"
  fi
  if [ -n "${XSELD_PID:-}" ]; then
    kill -TERM "$XSELD_PID"
  fi
  if [ -n "${XVFB_PID:-}" ]; then
    kill -TERM "$XVFB_PID"
  fi
  if [ -n "${DRIVER_PID:-}" ]; then
    kill -TERM "$DRIVER_PID"
  fi
  if [ -n "${X11VNC_PID:-}" ]; then
    kill -TERM "$X11VNC_PID"
  fi
  if [ -n "${PULSE_PID:-}" ]; then
    kill -TERM "$PULSE_PID"
  fi
}

trap clean SIGINT SIGTERM

/usr/bin/fileserver &
FILESERVER_PID=$!
/usr/bin/xseld &
XSELD_PID=$!

pulseaudio --start --exit-idle-time=-1
pactl load-module module-native-protocol-tcp
PULSE_PID=$(ps --no-headers -C pulseaudio -o pid | awk '{print $1}')

/usr/bin/xvfb-run -l -n "$DISPLAY_NUM" -s "-ac -screen 0 $SCREEN_RESOLUTION -noreset -listen tcp" /usr/bin/fluxbox -display "$DISPLAY" -log /dev/null 2>/dev/null &
XVFB_PID=$!
until wmctrl -m >/dev/null 2>&1; do
  echo Waiting X server...
  sleep 0.1
done

if [ "${ENABLE_VNC:-false}" == true ]; then
  x11vnc -display "$DISPLAY" -passwd "${VNC_PASSWORD:-selebrow}" -shared -forever -loop500 -rfbport 5900 -rfbportv6 5900 -logfile /dev/null &
  X11VNC_PID=$!
fi

# shellcheck disable=SC2086
/usr/bin/geckodriver --host 0.0.0.0 --port=4444 $DRIVER_ARGS &
DRIVER_PID=$!

wait
