# Introduction

TODO

## Supported hardware

The following table summarizes currently supported SoCs and boards.

| SoC          | Board                                                                                                                                                                                | SoC package                                                               | Board package                                                                        |
|--------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| NXP i.MX6UL  | [USB armory Mk II LAN](https://github.com/usbarmory/usbarmory/wiki)                                                                                                                  | [imx6ul](https://github.com/usbarmory/tamago/tree/master/soc/nxp/imx6ul)  | [usbarmory/mk2](https://github.com/usbarmory/tamago/tree/master/board/usbarmory)      |

## Secure Boot

On secure booted systems the `imx_signed` target should be used (instead of the unsigned `imx` one) with the relevant
[`HAB_KEYS`](https://github.com/usbarmory/usbarmory/wiki/Secure-boot-(Mk-II)) set.

## Kernel signing

To maintain the chain of trust the target kernel must be signed, to this end
the `OS_PUBLIC_KEY1` and `OS_PUBLIC_KEY2` environment variables must be set to the
path of either [signify](https://man.openbsd.org/signify) or
[minisign](https://jedisct1.github.io/minisign/) authentication keys, while
compiling.

Example key generation (signify):

```
signify -G -p armored-witness-os-1.pub -s armored-witness-os-1.sec
```

Example key generation (minisign):

```
minisign -G -p armored-witness-os-1.pub -s armored-witness-os-1.sec
```

Example signature generation (signify):

```
signify -S -s armored-witness-os-1.sec -m kernel -x kernel.sig1
```

Example signature generation (minisign):

```
minisign -S -s armored-witness-os-1.sec -m kernel -x kernel.sig1
```

## Compiling

Build the [TamaGo compiler](https://github.com/usbarmory/tamago-go)
(or use the [latest binary release](https://github.com/usbarmory/tamago-go/releases/latest)):

```
wget https://github.com/usbarmory/tamago-go/archive/refs/tags/latest.zip
unzip latest.zip
cd tamago-go-latest/src && ./all.bash
cd ../bin && export TAMAGO=`pwd`/go
```

Example compilation with embedded keys and secure boot:

```
git clone https://github.com/transparency-dev/armored-witness-boot && cd armored-witness-boot
make PUBLIC_KEY1=armored-witness-boot-1.pub PUBLIC_KEY2=armored-witness-boot-2.pub HAB_KEYS=sb_keys imx_signed
```

## Installing

> :warning: this is a work in progress

The `armored-witness-image` command line utility allows to create an image with
a valid bootloader configuration and kernel image.

You can automatically download, compile and install the utility, under your
GOPATH, as follows:

```
go install github.com/transparency-dev/armored-witness-boot/cmd/armored-witness-image@latest
```

Alternatively you can manually compile it from source:

```
git clone https://github.com/transparency-dev/armored-witness-boot
cd armores-witness-boot && make armored-witness-image
```

The utility output is meant to be flashed on the device using the
[armory-ums](https://github.com/usbarmory/armory-ums) firmware, loaded using
[armory-boot-usb](https://github.com/usbarmory/armory-boot/tree/master/cmd/armory-boot-usb).

The following example illustrates the required steps to flash the bootloader
and, separately, its configuration+kernel target.

```
# download armory-boot-usb utility
go install github.com/usbarmory/armory-boot/cmd/armory-boot-usb@latest

# download armory-ums firmware
wget https://github.com/usbarmory/armory-ums/releases/download/v20201102/armory-ums.imx

# download armored-witness-boot firmware
wget https://github.com/transparency-dev/armored-witness-boot/releases/download/v20201102/armored-witness-boot.imx

# download armored-witness-image utility
go install github.com/transparency-dev/armored-witness-boot/cmd/armored-witness-image@latest

# download armored-witness-os firmware
wget https://github.com/transparency-dev/armores-witness/releases/download/v20201102/armored-witness-os

# download armored-witness signatures
wget https://github.com/transparency-dev/armored-witness-os/releases/download/v20201102/armored-witness-os.sig1
wget https://github.com/transparency-dev/armored-witness-os/releases/download/v20201102/armored-witness.sig2

# create armored-witness-boot configuration
armored-witness-image -k trusted_os.elf -1 trusted_os.sig1 -2 trusted_os.sig2 -o trusted_os.bin

# load armory-ums on target
armory-boot-usb -i armory-ums.imx

# wait for mass storage device detection

# flash bootloader (verify target using `dmesg`)
dd if=armored-witness-boot.imx of=$TARGET_DEV bs=512 seek=2 conv=fsync

# flash configuration+kernel
dd if=armored-witnes-os.bin of=$TARGET_DEV bs=512 seek=20480 conv=fsync
```

## LED status

The [USB armory Mk II](https://github.com/usbarmory/usbarmory/wiki) LEDs
are used, in sequence, as follows:

| Boot sequence                   | Blue | White |
|---------------------------------|------|-------|
| 0. initialization               | off  | off   |
| 1. boot media detected          | on   | off   |
| 2. kernel verification complete | on   | on    |
| 3. jumping to kernel image      | off  | off   |
