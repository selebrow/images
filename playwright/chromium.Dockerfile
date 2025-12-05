ARG BASE_IMAGE=ghcr.io/selebrow/base/ubuntu-noble
ARG BASE_IMAGE_TAG=latest

FROM ${BASE_IMAGE}:${BASE_IMAGE_TAG}

USER root
COPY rootfs/ /

RUN export DEBIAN_FRONTEND=noninteractive && \
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get update && \
    apt-get install -y --no-install-recommends nodejs p11-kit libnss3 libxss1 libasound2t64 \
        libatk-bridge2.0-0t64 libgbm1 xdg-utils wget libu2f-udev libvulkan1 unzip && \
    # make chromium use system-wide trust store
    ln -sf /usr/lib/*64-linux-gnu/pkcs11/p11-kit-trust.so /usr/lib/*64-linux-gnu/libnssckbi.so && \
    apt-get clean && rm -rf /tmp/* && rm -Rf /var/lib/apt/lists/*

USER ${SB_USER}
WORKDIR ${SB_USER_HOME}

ARG PLAYWRIGHT_VERSION=1.53.2
RUN PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1 npm install playwright@$PLAYWRIGHT_VERSION &&  \
    npm install playwright-chromium@$PLAYWRIGHT_VERSION

ENTRYPOINT [ "dumb-init", "--", "/entrypoint.sh" ]
