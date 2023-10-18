# Introduction

TODO

## Supported hardware

The following table summarizes currently supported SoCs and boards.

| SoC          | Board                                                               | SoC package                                                              | Board package                                                                    |
|--------------|---------------------------------------------------------------------|--------------------------------------------------------------------------|----------------------------------------------------------------------------------|
| NXP i.MX6UL  | [USB armory Mk II LAN](https://github.com/usbarmory/usbarmory/wiki) | [imx6ul](https://github.com/usbarmory/tamago/tree/master/soc/nxp/imx6ul) | [usbarmory/mk2](https://github.com/usbarmory/tamago/tree/master/board/usbarmory) |
| NXP i.MX6ULL | [USB armory Mk II](https://github.com/usbarmory/usbarmory/wiki)     | [imx6ul](https://github.com/usbarmory/tamago/tree/master/soc/nxp/imx6ul) | [usbarmory/mk2](https://github.com/usbarmory/tamago/tree/master/board/usbarmory) |

## Secure Boot

On secure booted systems the `imx_signed` target should be used (instead of the unsigned `imx` one) with the relevant
[`HAB_KEYS`](https://github.com/usbarmory/usbarmory/wiki/Secure-boot-(Mk-II)) set.

## Kernel authentication

For an overview of the firmware authentication process please see
<https://github.com/transparency-dev/armored-witness/tree/main/docs/firmware_auth.md>.

To maintain the chain of trust, the bootloader authenticates the kernel before
executing it.

## Firmware transparency

All ArmoredWitness firmware artefacts need to be added to a firmware transparency log,
including the bootloader.

The provided `Makefile` has support for maintaining a local firmware transparency
log on disk. This is intended to be used for development only.

In order to use this functionality, a log key pair can be generated with the
following command:

```bash
$ go run github.com/transparency-dev/serverless-log/cmd/generate_keys@HEAD \
  --key_name="DEV-Log" \
  --out_priv=armored-witness-log.sec \
  --out_pub=armored-witness-log.pub
```

## Compiling

### Building the compiler

Build the [TamaGo compiler](https://github.com/usbarmory/tamago-go)
(or use the [latest binary release](https://github.com/usbarmory/tamago-go/releases/latest)):

```bash
wget https://github.com/usbarmory/tamago-go/archive/refs/tags/latest.zip
unzip latest.zip
cd tamago-go-latest/src && ./all.bash
cd ../bin && export TAMAGO=`pwd`/go
```

### Building the bootloader

Ensure the following environment variables are set:

| Variable            | Description
|---------------------|------------
| `BOOT_PRIVATE_KEY`  | Path to bootloader firmware signing key. Used by the Makefile to sign the bootloader.
| `OS_PUBLIC_KEY1`    | Path to OS firmware verification key 1. Embedded into the bootloader to verify the OS at run-time.
| `OS_PUBLIC_KEY2`    | Path to OS firmware verification key 2. Embedded into the bootloader to verify the OS at run-time.
| `LOG_PUBLIC_KEY`    | Path to log verification key. Embedded into the bootloader to verify at run-time that the OS is correctly logged.
| `LOG_ORIGIN`        | FT log origin string. Embedded into the bootloader to verify OS firmware transparency.
| `LOG_PRIVATE_KEY`   | Path to log signing key. Used by Makefile to add the new bootloader firmware to the local dev log.
| `DEV_LOG_DIR`       | Path to directory in which to store the dev FT log files.

Example compilation with embedded keys, ready for installation with the `provision` tool:

```bash
# Variables as above already exported.
make imx manifest log_boot
```

Example compilation with embedded keys and secure boot:

```bash
git clone https://github.com/transparency-dev/armored-witness-boot && cd armored-witness-boot
make OS_PUBLIC_KEY1=armored-witness-boot-1.pub OS_PUBLIC_KEY2=armored-witness-boot-2.pub HAB_KEYS=sb_keys imx_signed
```

### Encrypted RAM support

Only on i.MX6UL P/Ns, `BEE=1` can be set to enable AES CTR encryption for all
external RAM using TamaGo [bee package](https://pkg.go.dev/github.com/usbarmory/tamago/soc/nxp/bee).

## Installing

Installing the various firmware images onto the device can be accomplished using the
[provision](https://github.com/transparency-dev/armored-witness/tree/main/cmd/provision)
tool.

## LED status

The [USB armory Mk II](https://github.com/usbarmory/usbarmory/wiki) LEDs
are used, in sequence, as follows:

| Boot sequence                   | Blue | White |
|---------------------------------|------|-------|
| 0. initialization               | off  | off   |
| 1. boot media detected          | on   | off   |
| 2. kernel verification complete | on   | on    |
| 3. jumping to kernel image      | off  | off   |
