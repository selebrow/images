ARG BASE_IMAGE=ghcr.io/selebrow/base/ubuntu-noble
ARG BASE_IMAGE_TAG=latest

FROM ${BASE_IMAGE}:${BASE_IMAGE_TAG}

USER root
COPY rootfs/ /

RUN export DEBIAN_FRONTEND=noninteractive && \
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get update && \
    apt-get install -y --no-install-recommends nodejs libwoff1 libopus0 libwebpdemux2 libepoxy0  \
        liblcms2-2 libenchant-2-2 libmanette-0.2-0 libsoup-3.0-0 libxkbcommon0 libgles2 \
        libgudev-1.0-0 libsecret-1-0 libhyphen0 libgdk-pixbuf2.0-0 libegl1 libxslt1.1 libevent-2.1-7  \
        libharfbuzz-icu0 libgstreamer-plugins-bad1.0-0 gstreamer1.0-plugins-good gstreamer1.0-libav  \
        libgstreamer-gl1.0-0 libva2 libatomic1 libavif16 libgtk-4-1 && \
    apt-get clean && rm -rf /tmp/* && rm -Rf /var/lib/apt/lists/*

USER ${SB_USER}
WORKDIR ${SB_USER_HOME}

ARG PLAYWRIGHT_VERSION=1.53.2
RUN PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1 npm install playwright@$PLAYWRIGHT_VERSION &&  \
    npm install playwright-webkit@$PLAYWRIGHT_VERSION

ENTRYPOINT [ "dumb-init", "--", "/entrypoint.sh" ]
