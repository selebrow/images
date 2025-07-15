#!/bin/bash

sudo() {
    if [ $(id -u) -eq 0 ]; then
        "$@"
    else
        command sudo "$@"
    fi
}

set -euxo pipefail
export DEBIAN_FRONTEND=noninteractive
wget -q https://dl.google.com/linux/linux_signing_key.pub -O- | sudo tee /etc/apt/trusted.gpg.d/google.asc > /dev/null
gpg -n -q --import --import-options import-show /etc/apt/trusted.gpg.d/google.asc | grep -q EB4C1BFD4F042F6DDDCCEC917721F63BD38B4796 || exit 1
echo "deb [signed-by=/etc/apt/trusted.gpg.d/google.asc] http://dl.google.com/linux/chrome/deb/ stable main" | sudo tee /etc/apt/sources.list.d/google-chrome.list > /dev/null
echo '
Package: *
Pin: origin dl.google.com
Pin-Priority: 1000
' | sudo tee /etc/apt/preferences.d/google > /dev/null
sudo apt-get update > /dev/null
ver=$(apt-cache policy google-chrome-stable | grep "Candidate:" | awk '{print $2}')
sudo apt-get download google-chrome-stable=$ver >/dev/null
mv google-chrome-stable*${ver}*.deb google-chrome-stable.deb
echo $ver
