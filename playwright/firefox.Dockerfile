ARG BASE_IMAGE=ghcr.io/selebrow/base/ubuntu-noble
ARG BASE_IMAGE_TAG=latest

FROM ${BASE_IMAGE}:${BASE_IMAGE_TAG}

USER root
COPY rootfs/ /

RUN export DEBIAN_FRONTEND=noninteractive && \
    curl -fsSL https://deb.nodesource.com/setup_18.x | bash - && \
    apt-get update && \
    apt-get install -y --no-install-recommends nodejs p11-kit libasound2t64 libx11-xcb1 \
        libdbus-glib-1-2 libxt6 libgtk-3-0 ffmpeg && \
    apt-get clean && rm -rf /tmp/* && rm -Rf /var/lib/apt/lists/*

USER ${SB_USER}
WORKDIR ${SB_USER_HOME}

ARG PLAYWRIGHT_VERSION=1.51.1
RUN PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1 npm install playwright@$PLAYWRIGHT_VERSION &&  \
    npm install playwright-firefox@$PLAYWRIGHT_VERSION && \
    # make firefox use system-wide trust store
    FIREFOX_REVISION=$(jq -r '.browsers[] | select(.name == "firefox") | .revision' <node_modules/playwright-core/browsers.json) && \
    ln -sf /usr/lib/*64-linux-gnu/pkcs11/p11-kit-trust.so .cache/ms-playwright/firefox-${FIREFOX_REVISION}/firefox/libnssckbi.so

ENTRYPOINT [ "dumb-init", "--", "/entrypoint.sh" ]
