FROM golang:1.22-bookworm

ARG TAMAGO_VERSION
ARG LOG_ORIGIN
ARG LOG_PUBLIC_KEY
ARG OS_PUBLIC_KEY1
ARG OS_PUBLIC_KEY2
ARG GIT_SEMVER_TAG
# Build environment variables. In addition to routing these through to the make
# command, they MUST also be committed to in the manifest.
ARG BEE
ARG CONSOLE

# Install dependencies.
RUN apt-get update && apt-get install -y make
RUN apt-get install -y wget
RUN apt-get install -y binutils-arm-none-eabi
RUN apt-get install -y u-boot-tools

RUN wget "https://github.com/usbarmory/tamago-go/releases/download/tamago-go${TAMAGO_VERSION}/tamago-go${TAMAGO_VERSION}.linux-amd64.tar.gz"
RUN tar -xvf "tamago-go${TAMAGO_VERSION}.linux-amd64.tar.gz" -C /

WORKDIR /build

COPY . .

# Set Tamago path for Make rule.
ENV TAMAGO=/usr/local/tamago-go/bin/go

# The Makefile expects verifiers to be stored in files, so do that.
RUN echo "${LOG_PUBLIC_KEY}" > /tmp/log.pub
RUN echo "${OS_PUBLIC_KEY1}" > /tmp/os1.pub
RUN echo "${OS_PUBLIC_KEY2}" > /tmp/os2.pub

# Firmware transparency parameters for output binary.
ENV LOG_ORIGIN=${LOG_ORIGIN} \
    LOG_PUBLIC_KEY="/tmp/log.pub" \
    OS_PUBLIC_KEY1="/tmp/os1.pub" \
    OS_PUBLIC_KEY2="/tmp/os2.pub" \
    GIT_SEMVER_TAG=${GIT_SEMVER_TAG} \
    BEE=${BEE} \
    CONSOLE=${CONSOLE}

RUN make imx
