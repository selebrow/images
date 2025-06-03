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
wget -q https://packages.mozilla.org/apt/repo-signing-key.gpg -O- | sudo tee /etc/apt/keyrings/packages.mozilla.org.asc > /dev/null
gpg -n -q --import --import-options import-show /etc/apt/keyrings/packages.mozilla.org.asc | grep -q 35BAA0B33E9EB396F59CA838C0BA5CE6DC6315A3 || exit 1
echo "deb [signed-by=/etc/apt/keyrings/packages.mozilla.org.asc] https://packages.mozilla.org/apt mozilla main" | sudo tee -a /etc/apt/sources.list.d/mozilla.list > /dev/null
echo '
Package: *
Pin: origin packages.mozilla.org
Pin-Priority: 1000
' | sudo tee /etc/apt/preferences.d/mozilla > /dev/null
sudo apt-get update >/dev/null
ver=$(apt-cache show firefox | grep "Version:" | grep -v snap | head -1 | awk '{print $2}')
sudo apt-get download firefox=$ver >/dev/null
mv firefox*${ver}*.deb firefox.deb
echo $ver
