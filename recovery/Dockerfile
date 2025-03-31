FROM golang:1.24.1-bookworm

ARG TAMAGO_VERSION
# Must be a valid commit hash for https://github.com/usbarmory/armory-ums/
ARG ARMORY_UMS_VERSION=850baf54809bd29548d6f817933240043400a4e1

# Install dependencies.
RUN apt-get update && apt-get install -y make wget u-boot-tools binutils-arm-none-eabi

RUN wget "https://github.com/usbarmory/tamago-go/releases/download/tamago-go${TAMAGO_VERSION}/tamago-go${TAMAGO_VERSION}.linux-amd64.tar.gz"
RUN tar -xvf "tamago-go${TAMAGO_VERSION}.linux-amd64.tar.gz" -C /
# Set Tamago path for Make rule.
ENV TAMAGO=/usr/local/tamago-go/bin/go

WORKDIR /build

RUN git clone https://github.com/usbarmory/armory-ums.git
RUN cd armory-ums && git reset --hard ${ARMORY_UMS_VERSION} && \
    REPRODUCIBLE=1 make imx && \
    git rev-parse --verify HEAD > armory-ums.imx.git-commit
