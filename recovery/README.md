# Recovery

The Dockerfile in this directory is used for building an image to be used as a
recovery/provisioning tool for the armored witness.

We currently use the firmware in the github.com/usbarmory/armory-ums repo as
the recovery tool.

While that repo does offer prebuilt binary releases, we rebuild from scratch
here so we can be sure about which TamaGo toolchain version is used, etc.
