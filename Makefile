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

BUILD_EPOCH ?= $(shell /bin/date -u "+%s")
BUILD_DATE ?= $(shell /bin/date -u "+%Y-%m-%d %H:%M:%S")
BUILD_TAGS = linkramsize,linkramstart,linkprintk
REV = $(shell git rev-parse --short HEAD 2> /dev/null)
GIT_SEMVER_TAG ?= $(shell (git describe --tags --exact-match --match 'v*.*.*' 2>/dev/null || git describe --match 'v*.*.*' --tags 2>/dev/null || git describe --tags 2>/dev/null || echo -n v0.0.${BUILD_EPOCH}+`git rev-parse HEAD`) | tail -c +2 )
LOG_VERIFIER = $(shell test ${LOG_PUBLIC_KEY} && cat ${LOG_PUBLIC_KEY})
OS_VERIFIERS = [\"$(shell test ${OS_PUBLIC_KEY1} && cat ${OS_PUBLIC_KEY1})\", \"$(shell test ${OS_PUBLIC_KEY2} && cat ${OS_PUBLIC_KEY2})\"]

TAMAGO_SEMVER = $(shell [ -n "${TAMAGO}" -a -x "${TAMAGO}" ] && ${TAMAGO} version | sed 's/.*go\([0-9]\.[0-9]*\.[0-9]*\).*/\1/')
MINIMUM_TAMAGO_VERSION=1.22.6

SHELL = /bin/bash

ifeq ("${BEE}","1")
	BUILD_TAGS := ${BUILD_TAGS},bee
endif

ifeq ("${CONSOLE}","on")
	BUILD_TAGS := ${BUILD_TAGS},console
endif

APP := armored-witness-boot
CMD := armored-witness-image
GOENV := GO_EXTLINK_ENABLED=0 CGO_ENABLED=0 GOOS=tamago GOARM=7 GOARCH=arm
TEXT_START := 0x90010000 # ramStart (defined in imx6/imx6ul/memory.go) + 0x10000
TAMAGOFLAGS := -tags ${BUILD_TAGS} -trimpath -buildvcs=false -buildmode=exe \
	-ldflags "-s -w -T $(TEXT_START) -E _rt0_arm_tamago -R 0x1000 \
			  -X 'main.Revision=${REV}' -X 'main.Version=${GIT_SEMVER_TAG}' \
			  -X 'main.OSLogOrigin=${LOG_ORIGIN}' \
			  -X 'main.OSLogVerifier=${LOG_VERIFIER}' \
			  -X 'main.OSManifestVerifiers=${OS_VERIFIERS}'"
GOFLAGS := -trimpath -buildvcs=false -buildmode=exe -ldflags "-s -w"

QEMU ?= qemu-system-arm -machine mcimx6ul-evk -cpu cortex-a7 -m 512M \
        -nographic -monitor none -serial null -serial stdio \
        -net nic,model=imx.enet,netdev=net0 -netdev tap,id=net0,ifname=tap0,script=no,downscript=no \
        -semihosting

.PHONY: clean qemu qemu-gdb

#### primary targets ####

all: $(APP)

imx: $(APP).imx
manifest: $(APP)_manifest

imx_signed: $(APP)-signed.imx $(APP)_manifest

elf: $(APP)

$(CMD):
	@if [ "${TAMAGO}" != "" ]; then \
		${TAMAGO} build $(GOFLAGS) cmd/$(CMD)/*.go; \
	else \
		go build $(GOFLAGS) cmd/$(CMD)/*.go; \
	fi

## log_initialise initialises the log stored under ${LOG_STORAGE_DIR}.
log_initialise:
	echo "(Re-)initialising log at ${LOG_STORAGE_DIR}"
	go run github.com/transparency-dev/serverless-log/cmd/integrate@a56a93b5681e5dc231882ac9de435c21cb340846 \
		--storage_dir=${LOG_STORAGE_DIR} \
		--origin=${LOG_ORIGIN} \
		--private_key=${LOG_PRIVATE_KEY} \
		--public_key=${LOG_PUBLIC_KEY} \
		--initialise

check_log:
	@if [ "${LOG_PRIVATE_KEY}" == "" -o "${LOG_PUBLIC_KEY}" == "" ]; then \
		@echo "You need to set LOG_PRIVATE_KEY and LOG_PUBLIC_KEY variables"; \
		exit 1; \
	fi
	@if [ "${DEV_LOG_DIR}" == "" ]; then \
		@echo "You need to set the DEV_LOG_DIR variable"; \
		exit 1; \
	fi

## log_boot adds the manifest.json file created during the build to the dev FT log.
log_boot: LOG_STORAGE_DIR=$(DEV_LOG_DIR)/log
log_boot: LOG_ARTEFACT_DIR=$(DEV_LOG_DIR)/artefacts
log_boot: ARTEFACT_HASH=$(shell sha256sum ${CURDIR}/${APP}.imx | cut -f1 -d" ")
log_boot: check_log
	@if [ ! -f ${LOG_STORAGE_DIR}/checkpoint ]; then \
		make log_initialise LOG_STORAGE_DIR="${LOG_STORAGE_DIR}" ; \
	fi
	go run github.com/transparency-dev/serverless-log/cmd/sequence@a56a93b5681e5dc231882ac9de435c21cb340846 \
		--storage_dir=${LOG_STORAGE_DIR} \
		--origin=${LOG_ORIGIN} \
		--public_key=${LOG_PUBLIC_KEY} \
		--entries=${CURDIR}/${APP}_manifest
	-go run github.com/transparency-dev/serverless-log/cmd/integrate@a56a93b5681e5dc231882ac9de435c21cb340846 \
		--storage_dir=${LOG_STORAGE_DIR} \
		--origin=${LOG_ORIGIN} \
		--private_key=${LOG_PRIVATE_KEY} \
		--public_key=${LOG_PUBLIC_KEY}
	@mkdir -p ${LOG_ARTEFACT_DIR}
	cp ${CURDIR}/${APP}.imx ${LOG_ARTEFACT_DIR}/${ARTEFACT_HASH}


## log_recovery creates a manifest for a defined version of the armory-ums image, and stores it
## in the local dev FT log.
## See https://github.com/usbarmory/armory-ums/releases
log_recovery: ARMORY_UMS_RELEASE=v20231018
log_recovery: ARMORY_UMS_GIT_TAG="0.0.0-incompatible+${ARMORY_UMS_RELEASE}" # Workaround for semver format requirement.
log_recovery: LOG_STORAGE_DIR=$(DEV_LOG_DIR)/log
log_recovery: LOG_ARTEFACT_DIR=$(DEV_LOG_DIR)/artefacts
log_recovery: TAMAGO_SEMVER=$(shell ${TAMAGO} version | sed 's/.*go\([0-9]\.[0-9]*\.[0-9]*\).*/\1/')
log_recovery: ARTEFACT_HASH=$(shell sha256sum ${CURDIR}/armory-ums.imx | cut -f1 -d" ")
log_recovery: check_log
	@if [ "${RECOVERY_PRIVATE_KEY}" == "" ]; then \
		@echo "You need to set RECOVERY_PRIVATE_KEY variable"; \
		exit 1; \
	fi
	docker build -t armory-ums-build -f recovery/Dockerfile --build-arg=TAMAGO_VERSION=${TAMAGO_SEMVER} --build-arg=ARMORY_UMS_VERSION=${ARMORY_UMS_RELEASE} --network=host  recovery/
	docker create --name au-build armory-ums-build
	docker cp au-build:/build/armory-ums/armory-ums.imx .
	docker cp au-build:/build/armory-ums/armory-ums.imx.git-commit .
	docker rm -v au-build

	@if [ ! -f ${LOG_STORAGE_DIR}/checkpoint ]; then \
		make log_initialise LOG_STORAGE_DIR="${LOG_STORAGE_DIR}" ; \
	fi
	go run github.com/transparency-dev/armored-witness/cmd/manifest@561c0b09a2cc48877a8c9e59c3fbf7ffc81cdd4d \
		create \
		--git_tag=${ARMORY_UMS_GIT_TAG} \
		--git_commit_fingerprint=$$(cat armory-ums.imx.git-commit) \
		--firmware_file=${CURDIR}/armory-ums.imx \
		--firmware_type=RECOVERY \
		--private_key_file=${RECOVERY_PRIVATE_KEY} \
		--tamago_version=${TAMAGO_SEMVER} \
		--output_file=${CURDIR}/armory-ums_manifest

	go run github.com/transparency-dev/serverless-log/cmd/sequence@a56a93b5681e5dc231882ac9de435c21cb340846 \
		--storage_dir=${LOG_STORAGE_DIR} \
		--origin=${LOG_ORIGIN} \
		--public_key=${LOG_PUBLIC_KEY} \
		--entries=${CURDIR}/armory-ums_manifest
	-go run github.com/transparency-dev/serverless-log/cmd/integrate@a56a93b5681e5dc231882ac9de435c21cb340846 \
		--storage_dir=${LOG_STORAGE_DIR} \
		--origin=${LOG_ORIGIN} \
		--private_key=${LOG_PRIVATE_KEY} \
		--public_key=${LOG_PUBLIC_KEY}
	@mkdir -p ${LOG_ARTEFACT_DIR}
	cp ${CURDIR}/armory-ums.imx ${LOG_ARTEFACT_DIR}/${ARTEFACT_HASH}


#### utilities ####

check_env:
	@if [ "${LOG_ORIGIN}" == "" ]; then \
		echo 'You need to set the LOG_ORIGIN variable'; \
		exit 1; \
	fi
	@if [ "${LOG_PUBLIC_KEY}" == "" ] || [ ! -f "${LOG_PUBLIC_KEY}" ]; then \
		echo 'You need to set the LOG_PUBLIC_KEY variable to a valid note verifier key path'; \
		exit 1; \
	fi
	@if [ "${OS_PUBLIC_KEY1}" == "" ] || [ ! -f "${OS_PUBLIC_KEY1}" ]; then \
		echo 'You need to set the OS_PUBLIC_KEY1 variable to a valid note verifier key path'; \
		exit 1; \
	fi
	@if [ "${OS_PUBLIC_KEY2}" == "" ] || [ ! -f "${OS_PUBLIC_KEY2}" ]; then \
		echo 'You need to set the OS_PUBLIC_KEY2 variable to a valid note verifier key path'; \
		exit 1; \
	fi

check_tamago:
	@if [ "${TAMAGO}" == "" ] || [ ! -f "${TAMAGO}" ]; then \
		echo 'You need to set the TAMAGO variable to a compiled version of https://github.com/usbarmory/tamago-go'; \
		exit 1; \
	fi
	@if [ "$(shell printf '%s\n' ${MINIMUM_TAMAGO_VERSION} ${TAMAGO_SEMVER} | sort -V | head -n1 )" != "${MINIMUM_TAMAGO_VERSION}" ]; then \
		echo "You need TamaGo >= ${MINIMUM_TAMAGO_VERSION}, found ${TAMAGO_SEMVER}" ; \
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
	@rm -fr $(APP) $(APP).bin $(APP).imx $(APP)-signed.imx $(APP).csf $(APP).dcd $(CMD) $(APP)_manifest

qemu: $(APP)
	$(QEMU) -kernel $(CURDIR)/armored-witness-boot

qemu-gdb: TAMAGOFLAGS := $(TAMAGOFLAGS:-w=)
qemu-gdb: TAMAGOFLAGS := $(TAMAGOFLAGS:-s=)
qemu-gdb: $(APP)
	$(QEMU) -kernel $(CURDIR)/armored-witness-boot -S -s

#### dependencies ####

$(APP): check_tamago check_env
	$(GOENV) $(TAMAGO) build $(TAMAGOFLAGS) -o ${APP}
	sha256sum $(APP)

$(APP).dcd: check_tamago
$(APP).dcd: GOMODCACHE=$(shell ${TAMAGO} env GOMODCACHE)
$(APP).dcd: TAMAGO_PKG=$(shell grep "github.com/usbarmory/tamago v" go.mod | awk '{print $$1"@"$$2}')
$(APP).dcd: dcd

$(APP).bin: CROSS_COMPILE=arm-none-eabi-
$(APP).bin: $(APP)
	$(CROSS_COMPILE)objcopy --enable-deterministic-archives \
	    -j .text -j .rodata -j .shstrtab -j .typelink \
	    -j .itablink -j .gopclntab -j .go.buildinfo -j .noptrdata -j .data \
	    -j .bss --set-section-flags .bss=alloc,load,contents \
	    -j .noptrbss --set-section-flags .noptrbss=alloc,load,contents \
	    $(APP) -O binary $(APP).bin
	sha256sum $(APP).bin

$(APP).imx: SOURCE_DATE_EPOCH=0
$(APP).imx: $(APP).bin $(APP).dcd
	echo "## disabling TZASC bypass in DCD for pre-DDR initialization ##"; \
	chmod 644 $(APP).dcd; \
	echo "DATA 4 0x020e4024 0x00000001  # TZASC_BYPASS" >> $(APP).dcd; \
	mkimage -v -n $(APP).dcd -T imximage -e $(TEXT_START) -d $(APP).bin $(APP).imx
	sha256sum $(APP).imx
	# Copy entry point from ELF file
	dd if=$(APP) of=$(APP).imx bs=1 count=4 skip=24 seek=4 conv=notrunc
	sha256sum $(APP).imx

$(APP)_manifest: imx
	@if [ "${BOOT_PRIVATE_KEY}" == "" ]; then \
		echo 'You need to set the BOOT_PRIVATE_KEY variable to a valid signing key path'; \
		exit 1; \
	fi

	# Create manifest
	@echo ---------- Manifest --------------
	go run github.com/transparency-dev/armored-witness/cmd/manifest@561c0b09a2cc48877a8c9e59c3fbf7ffc81cdd4d \
		create \
		--git_tag=${GIT_SEMVER_TAG} \
		--git_commit_fingerprint="${REV}" \
		--firmware_file=${CURDIR}/$(APP).imx \
		--firmware_type=BOOTLOADER \
		--private_key_file=${BOOT_PRIVATE_KEY} \
		--tamago_version=${TAMAGO_SEMVER} \
		--output_file=${CURDIR}/${APP}_manifest
	@echo ----------------------------------

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
