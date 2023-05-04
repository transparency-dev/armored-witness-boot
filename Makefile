# Copyright 2022 The Armored Witness Boot authors. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

BUILD_USER ?= $(shell whoami)
BUILD_HOST ?= $(shell hostname)
BUILD_DATE ?= $(shell /bin/date -u "+%Y-%m-%d %H:%M:%S")
BUILD_TAGS = linkramsize,linkramstart
BUILD = ${BUILD_USER}@${BUILD_HOST} on ${BUILD_DATE}
REV = $(shell git rev-parse --short HEAD 2> /dev/null)
PUBLIC_KEYS = [\"$(shell test ${OS_PUBLIC_KEY1} && tail -n 1 ${OS_PUBLIC_KEY1})\", \"$(shell test ${OS_PUBLIC_KEY2} && tail -n 1 ${OS_PUBLIC_KEY2})\"]

SHELL = /bin/bash

ifeq ("${CONSOLE}","on")
	BUILD_TAGS := ${BUILD_TAGS},console
endif

APP := armored-witness-boot
CMD := armored-witness-image
GOENV := GO_EXTLINK_ENABLED=0 CGO_ENABLED=0 GOOS=tamago GOARM=7 GOARCH=arm
TEXT_START := 0x90010000 # ramStart (defined in imx6/imx6ul/memory.go) + 0x10000
TAMAGOFLAGS := -tags ${BUILD_TAGS} -trimpath -ldflags "-s -w -T $(TEXT_START) -E _rt0_arm_tamago -R 0x1000 -X 'main.Build=${BUILD}' -X 'main.Revision=${REV}' -X 'main.PublicKeys=${PUBLIC_KEYS}'"
GOFLAGS := -trimpath -ldflags "-s -w"

.PHONY: clean

#### primary targets ####

all: $(APP)

imx: $(APP).imx

imx_signed: $(APP)-signed.imx

elf: $(APP)

$(CMD):
	@if [ "${TAMAGO}" != "" ]; then \
		${TAMAGO} build $(GOFLAGS) cmd/$(CMD)/*.go; \
	else \
		go build $(GOFLAGS) cmd/$(CMD)/*.go; \
	fi

#### utilities ####

check_env:
	@if [ "${OS_PUBLIC_KEY1}" == "" ] || [ ! -f "${OS_PUBLIC_KEY1}" ]; then \
		echo 'You need to set the OS_PUBLIC_KEY1 variable to a valid authentication key path'; \
		exit 1; \
	fi
	@if [ "${OS_PUBLIC_KEY2}" == "" ] || [ ! -f "${OS_PUBLIC_KEY2}" ]; then \
		echo 'You need to set the OS_PUBLIC_KEY2 variable to a valid authentication key path'; \
		exit 1; \
	fi

check_tamago:
	@if [ "${TAMAGO}" == "" ] || [ ! -f "${TAMAGO}" ]; then \
		echo 'You need to set the TAMAGO variable to a compiled version of https://github.com/usbarmory/tamago-go'; \
		exit 1; \
	fi

check_hab_keys:
	@if [ "${HAB_KEYS}" == "" ]; then \
		echo 'You need to set the HAB_KEYS variable to the path of secure boot keys'; \
		echo 'See https://github.com/usbarmory/usbarmory/wiki/Secure-boot-(Mk-II)'; \
		exit 1; \
	fi

dcd:
	echo $(GOMODCACHE)
	echo $(TAMAGO_PKG)
	cp -f $(GOMODCACHE)/$(TAMAGO_PKG)/board/usbarmory/mk2/imximage.cfg $(APP).dcd

clean:
	@rm -fr $(APP) $(APP).bin $(APP).imx $(APP)-signed.imx $(APP).csf $(APP).dcd $(CMD)

#### dependencies ####

$(APP): check_tamago check_env
	$(GOENV) $(TAMAGO) build $(TAMAGOFLAGS) -o ${APP}

$(APP).dcd: check_tamago
$(APP).dcd: GOMODCACHE=$(shell ${TAMAGO} env GOMODCACHE)
$(APP).dcd: TAMAGO_PKG=$(shell grep "github.com/usbarmory/tamago v" go.mod | awk '{print $$1"@"$$2}')
$(APP).dcd: dcd

$(APP).bin: CROSS_COMPILE=arm-none-eabi-
$(APP).bin: $(APP)
	$(CROSS_COMPILE)objcopy -j .text -j .rodata -j .shstrtab -j .typelink \
	    -j .itablink -j .gopclntab -j .go.buildinfo -j .noptrdata -j .data \
	    -j .bss --set-section-flags .bss=alloc,load,contents \
	    -j .noptrbss --set-section-flags .noptrbss=alloc,load,contents \
	    $(APP) -O binary $(APP).bin

$(APP).imx: $(APP).bin $(APP).dcd
	echo "## disabling TZASC bypass in DCD for pre-DDR initialization ##"; \
	chmod 644 $(APP).dcd; \
	echo "DATA 4 0x020e4024 0x00000001  # TZASC_BYPASS" >> $(APP).dcd; \
	mkimage -n $(APP).dcd -T imximage -e $(TEXT_START) -d $(APP).bin $(APP).imx
	# Copy entry point from ELF file
	dd if=$(APP) of=$(APP).imx bs=1 count=4 skip=24 seek=4 conv=notrunc

#### secure boot ####

$(APP)-signed.imx: check_hab_keys $(APP).imx
	${TAMAGO} install github.com/usbarmory/crucible/cmd/habtool
	$(shell ${TAMAGO} env GOPATH)/bin/habtool \
		-A ${HAB_KEYS}/CSF_1_key.pem \
		-a ${HAB_KEYS}/CSF_1_crt.pem \
		-B ${HAB_KEYS}/IMG_1_key.pem \
		-b ${HAB_KEYS}/IMG_1_crt.pem \
		-t ${HAB_KEYS}/SRK_1_2_3_4_table.bin \
		-x 1 \
		-i $(APP).imx \
		-o $(APP).csf && \
	cat $(APP).imx $(APP).csf > $(APP)-signed.imx
