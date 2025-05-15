ARG BASE_IMAGE=ghcr.io/selebrow/base/ubuntu-noble
ARG BASE_IMAGE_TAG=latest

FROM ${BASE_IMAGE}:${BASE_IMAGE_TAG}

USER root
COPY rootfs/ /

RUN export DEBIAN_FRONTEND=noninteractive && \
    chmod 755 /usr/bin/fileserver /usr/bin/xseld && \
    apt-get update && \
    apt-get install -y --no-install-recommends p11-kit libnss3 libxss1 libasound2t64 libatk-bridge2.0-0 libgbm1 \
            xdg-utils wget libu2f-udev libvulkan1 unzip && \
    # make chrome use system-wide trust store
    ln -sf /usr/lib/x86_64-linux-gnu/pkcs11/p11-kit-trust.so /usr/lib/x86_64-linux-gnu/libnssckbi.so && \
    apt-get clean && rm -rf /tmp/* && rm -Rf /var/lib/apt/lists/*

ARG CHROMEDRIVER_URL
ARG CHROME_URL

RUN export DEBIAN_FRONTEND=noninteractive && \
    curl -fSsL ${CHROMEDRIVER_URL} -o /tmp/chromedriver-linux64.zip && \
    unzip -j /tmp/chromedriver-linux64.zip chromedriver-linux64/chromedriver -d /usr/bin && chmod 755 /usr/bin/chromedriver && \
    chromedriver --version && \
    curl -fSsL ${CHROME_URL} -o /tmp/google-chrome-stable.deb && \
    dpkg -i /tmp/google-chrome-stable.deb && \
    sed -i -e 's@exec -a "$0" "$HERE/chrome"@& --no-sandbox --disable-gpu@' /opt/google/chrome/google-chrome && \
    chown root:root /opt/google/chrome/chrome-sandbox && chmod 4755 /opt/google/chrome/chrome-sandbox && \
    google-chrome --version && \
    rm -rf /tmp/*

COPY chrome/rootfs/ /

RUN chmod 755 /usr/bin/devtools

USER ${SB_USER}
WORKDIR ${SB_USER_HOME}


ENTRYPOINT [ "/entrypoint.sh" ]
