#!/bin/bash

set -euxo pipefail
URL="https://googlechromelabs.github.io/chrome-for-testing/LATEST_RELEASE_STABLE"
ver=$(curl -fsSL "$URL")
echo $ver
