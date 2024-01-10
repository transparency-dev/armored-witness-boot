FROM golang:1.21.3-bookworm

ARG TAMAGO_VERSION
# Pinned at tag [v20231018](https://github.com/usbarmory/armory-ums/releases/tag/v20231018)
ARG ARMORY_UMS_VERSION=74060722c9aa92bbdcf3725ed0d0be4ebe8f8687
ARG ARMORED_WITNESS_VERSION=228f2f6432babe1f1657e150ce0ca4a96ab394da

# Install dependencies.
RUN apt-get update && apt-get install -y make wget u-boot-tools binutils-arm-none-eabi

RUN wget "https://github.com/usbarmory/tamago-go/releases/download/tamago-go${TAMAGO_VERSION}/tamago-go${TAMAGO_VERSION}.linux-amd64.tar.gz"
RUN tar -xvf "tamago-go${TAMAGO_VERSION}.linux-amd64.tar.gz" -C /
# Set Tamago path for Make rule.
ENV TAMAGO=/usr/local/tamago-go/bin/go

WORKDIR /build

RUN git clone https://github.com/usbarmory/armory-ums.git
RUN cd armory-ums && git reset --hard ${ARMORY_UMS_VERSION} && \
    CROSS_COMPILE=arm-none-eabi- make imx && \
    git rev-parse --verify HEAD > armory-ums.imx.git-commit