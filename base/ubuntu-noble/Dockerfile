ARG BASE_IMAGE=ubuntu
ARG BASE_IMAGE_TAG=24.04
FROM ${BASE_IMAGE}:${BASE_IMAGE_TAG}

COPY rootfs/ /

RUN export DEBIAN_FRONTEND=noninteractive && \
    apt-get update && apt-get -y upgrade && \
    echo ttf-mscorefonts-installer msttcorefonts/accepted-mscorefonts-eula select true | debconf-set-selections && \
    apt-get install -y --no-install-recommends ca-certificates tzdata locales \
            gnupg curl jq libgtk-3-0 ttf-mscorefonts-installer fontconfig \
            fontconfig-config fonts-dejavu-core fonts-liberation fonts-ubuntu-console \
            fonts-wqy-zenhei fonts-thai-tlwg-ttf fonts-ipafont-mincho fonts-sahadeva \
            fonts-noto-unhinted fonts-noto-color-emoji libfontconfig1 libfontenc1 \
            libfreetype6 libxfont2 libxft2 libnss3-tools xfonts-base xfonts-encodings \
            xfonts-utils xvfb pulseaudio fluxbox x11vnc feh wmctrl libnss-wrapper xsel && \
    fc-cache -fv && \
        mkdir -p /var/lib/locales/supported.d/ && \
        grep -E '(en_US|ru_RU).UTF-8' /usr/share/i18n/SUPPORTED > /var/lib/locales/supported.d/all && \
        locale-gen && update-locale && \
        apt-get clean && rm -rf /var/lib/apt/lists/*

ENV SB_USER=selebrow
ENV SB_USER_HOME=/home/${SB_USER}

RUN addgroup --system ${SB_USER} && \
    adduser --system --home ${SB_USER_HOME} --uid 2048 --ingroup ${SB_USER} --disabled-password --shell /bin/bash ${SB_USER} && \
    adduser ${SB_USER} audio && adduser ${SB_USER} video && \
    mkdir -p ${SB_USER_HOME}/Downloads && chown -R ${SB_USER}:${SB_USER} ${SB_USER_HOME}

USER ${SB_USER}
WORKDIR ${SB_USER_HOME}

ENV LANG=en_US.UTF-8
ENV LANGUAGE=en_US:en
ENV LC_ALL=en_US.UTF-8
