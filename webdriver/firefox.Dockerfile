ARG BASE_IMAGE=ghcr.io/selebrow/base/ubuntu-noble
ARG BASE_IMAGE_TAG=latest

FROM ${BASE_IMAGE}:${BASE_IMAGE_TAG}

USER root
COPY rootfs/ /

RUN export DEBIAN_FRONTEND=noninteractive && \
    chmod 755 /usr/bin/fileserver /usr/bin/xseld && \
    apt-get update && \
    apt-get install -y --no-install-recommends p11-kit libavcodec60 libdbus-glib-1-2 && \
    apt-get clean && rm -rf /tmp/* && rm -Rf /var/lib/apt/lists/*

RUN --mount=type=bind,source=browser_data,target=/data \
    export DEBIAN_FRONTEND=noninteractive && \
    tar xzf /data/geckodriver-linux64.tar.gz -C /usr/bin && chmod 755 /usr/bin/geckodriver && \
    geckodriver --version && \
    dpkg -i /data/firefox.deb && \
    # make firefox use system-wide trust store
    ( ln -sf /usr/lib/x86_64-linux-gnu/pkcs11/p11-kit-trust.so /usr/lib/firefox/libnssckbi.so || ln -sf /usr/lib/x86_64-linux-gnu/pkcs11/p11-kit-trust.so /opt/firefox/libnssckbi.so ) && \
    firefox --version

COPY firefox/rootfs/ /

USER ${SB_USER}
WORKDIR ${SB_USER_HOME}


ENTRYPOINT [ "/entrypoint.sh" ]
