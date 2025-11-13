ARG BASE_IMAGE=ghcr.io/selebrow/base/ubuntu-noble
ARG BASE_IMAGE_TAG=latest

FROM ${BASE_IMAGE}:${BASE_IMAGE_TAG}

USER root
COPY rootfs/ /

RUN export DEBIAN_FRONTEND=noninteractive && \
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get update && \
    apt-get install -y --no-install-recommends nodejs p11-kit libnss3 libxss1 libasound2t64 libatk-bridge2.0-0 libgbm1 \
        ffmpeg xdg-utils wget libu2f-udev libvulkan1 unzip && \
    # make chrome using system-wide trust store
    ln -sf /usr/lib/x86_64-linux-gnu/pkcs11/p11-kit-trust.so /usr/lib/x86_64-linux-gnu/libnssckbi.so && \
    apt-get clean && rm -rf /tmp/* && rm -Rf /var/lib/apt/lists/*

RUN --mount=type=bind,source=browser_data,target=/data \
    export DEBIAN_FRONTEND=noninteractive && \
    dpkg -i /data/google-chrome-stable.deb

USER ${SB_USER}
WORKDIR ${SB_USER_HOME}

ARG PLAYWRIGHT_VERSION=1.53.1
RUN --mount=type=bind,source=browser_data,target=/data \
    export PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1 && \
    npm install playwright@${PLAYWRIGHT_VERSION} playwright-chromium@${PLAYWRIGHT_VERSION} && \
    CHROMIUM_REVISION=$(jq -r '.browsers[] | select(.name == "chromium") | .revision' <node_modules/playwright-core/browsers.json) && \
    mkdir -p ${SB_USER_HOME}/.cache/ms-playwright/chromium-${CHROMIUM_REVISION}/chrome-linux && \
    ln -s /opt/google/chrome/google-chrome ${SB_USER_HOME}/.cache/ms-playwright/chromium-${CHROMIUM_REVISION}/chrome-linux/chrome && \
    FFMPEG_REVISION=$(jq -r '.browsers[] | select(.name == "ffmpeg") | .revision' <node_modules/playwright-core/browsers.json) && \
    FFMPEG_CACHE_DIR=${SB_USER_HOME}/.cache/ms-playwright/ffmpeg-${FFMPEG_REVISION} && \
    mkdir -p $FFMPEG_CACHE_DIR && \
    ln -s /usr/bin/ffmpeg ${FFMPEG_CACHE_DIR}/ffmpeg-linux && \
    CHROME_HEADLESS_CACHE_DIR=${SB_USER_HOME}/.cache/ms-playwright/chromium_headless_shell-${CHROMIUM_REVISION} && \
    mkdir -p $CHROME_HEADLESS_CACHE_DIR && \
    unzip /data/chrome-headless-shell-linux64.zip -d ${CHROME_HEADLESS_CACHE_DIR} && \
    mv ${CHROME_HEADLESS_CACHE_DIR}/chrome-headless-shell-linux64 ${CHROME_HEADLESS_CACHE_DIR}/chrome-linux && \
    ln -s ${CHROME_HEADLESS_CACHE_DIR}/chrome-linux/chrome-headless-shell ${CHROME_HEADLESS_CACHE_DIR}/chrome-linux/headless_shell && \
    ${CHROME_HEADLESS_CACHE_DIR}/chrome-linux/headless_shell --version

ENTRYPOINT [ "dumb-init", "--", "/entrypoint.sh" ]
