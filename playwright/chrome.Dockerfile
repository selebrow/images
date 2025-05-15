ARG BASE_IMAGE=ghcr.io/selebrow/base/ubuntu-noble
ARG BASE_IMAGE_TAG=latest

FROM ${BASE_IMAGE}:${BASE_IMAGE_TAG}

USER root
COPY rootfs/ /

RUN export DEBIAN_FRONTEND=noninteractive && \
    curl -fsSL https://deb.nodesource.com/setup_18.x | bash - && \
    apt-get update && \
    apt-get install -y --no-install-recommends nodejs p11-kit libnss3 libxss1 libasound2t64 libatk-bridge2.0-0 libgbm1 \
            ffmpeg xdg-utils wget libu2f-udev libvulkan1 unzip dumb-init && \
    # make chrome using system-wide trust store
    ln -sf /usr/lib/x86_64-linux-gnu/pkcs11/p11-kit-trust.so /usr/lib/x86_64-linux-gnu/libnssckbi.so && \
    apt-get clean && rm -rf /tmp/* && rm -Rf /var/lib/apt/lists/*

ARG CHROME_URL
ARG CHROME_HEADLESS_URL

RUN export DEBIAN_FRONTEND=noninteractive && \
    curl -fSsL ${CHROME_URL} -o /tmp/google-chrome-stable.deb && \
    dpkg -i /tmp/google-chrome-stable.deb && \
    rm -rf /tmp/*

USER ${SB_USER}
WORKDIR ${SB_USER_HOME}

ARG PLAYWRIGHT_VERSION=1.51.1
RUN export PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1 && \
    npm install playwright@${PLAYWRIGHT_VERSION} playwright-chromium@${PLAYWRIGHT_VERSION} && \
    CHROMIUM_REVISION=$(jq -r '.browsers[] | select(.name == "chromium") | .revision' <node_modules/playwright-core/browsers.json) && \
    mkdir -p ${SB_USER_HOME}/.cache/ms-playwright/chromium-${CHROMIUM_REVISION}/chrome-linux && \
    ln -s /opt/google/chrome/google-chrome ${SB_USER_HOME}/.cache/ms-playwright/chromium-${CHROMIUM_REVISION}/chrome-linux/chrome && \
    FFMPEG_REVISION=$(jq -r '.browsers[] | select(.name == "ffmpeg") | .revision' <node_modules/playwright-core/browsers.json) && \
    FFMPEG_CACHE_DIR=${SB_USER_HOME}/.cache/ms-playwright/ffmpeg-${FFMPEG_REVISION} && \
    mkdir -p $FFMPEG_CACHE_DIR && \
    ln -s /usr/bin/ffmpeg ${FFMPEG_CACHE_DIR}/ffmpeg-linux && \
    CHROME_HEADLESS_CACHE_DIR=${SB_USER_HOME}/.cache/ms-playwright/chromium_headless_shell-${CHROMIUM_REVISION} && \
    curl -fSsL ${CHROME_HEADLESS_URL} -o /tmp/chrome-headless-shell-linux64.zip && \
    mkdir -p $CHROME_HEADLESS_CACHE_DIR && \
    unzip /tmp/chrome-headless-shell-linux64.zip -d ${CHROME_HEADLESS_CACHE_DIR} && \
    mv ${CHROME_HEADLESS_CACHE_DIR}/chrome-headless-shell-linux64 ${CHROME_HEADLESS_CACHE_DIR}/chrome-linux && \
    ln -s ${CHROME_HEADLESS_CACHE_DIR}/chrome-linux/chrome-headless-shell ${CHROME_HEADLESS_CACHE_DIR}/chrome-linux/headless_shell && \
    ${CHROME_HEADLESS_CACHE_DIR}/chrome-linux/headless_shell --version && \
    rm -f /tmp/chrome-headless-shell-linux64.zip

ENTRYPOINT [ "dumb-init", "--", "/entrypoint.sh" ]
