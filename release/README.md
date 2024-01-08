# Bootloader Release Process

## File structure

*   The Dockerfile found in the root of the repo builds an image which installs
    dependencies and compiles the bootloader with TamaGo. The version of
    TamaGo to use can be specified with the Docker
    [build arg](https://docs.docker.com/engine/reference/commandline/build/#build-arg)
    `TAMAGO_VERSION`.
*   Cloud Build triggers for the presubmit, continuous integration (CI)m and
    prod environments are defined on the Cloud Build yaml files in this
    directory.

## Build and Release Process

A
[Cloud Build trigger](https://cloud.google.com/build/docs/automating-builds/create-manage-triggers)
is defined by a yaml config file and is invoked when a new tag is published in
this repository.

The pipeline includes two main steps: building and making available the
bootloader imx, and writing the release metadata (Claimant Model Statement) to
the firmware transparency log.

1.  Cloud Build builds the bootloader builder Docker image and uploads the
    compiled bootloader imx file to a public Google Cloud Storage bucket.
1.  Cloud Build runs the
    [`manifest`](https://github.com/transparency-dev/armored-witness/tree/main/cmd/manifest)
    tool to construct the Claimant Model Statement with arguments specific to
    this release. It signs the Statement with the
    [`sign`](https://github.com/transparency-dev/armored-witness/tree/main/cmd/sign)
    tool and adds the resulting signed Statement as an entry to the public
    firmware transparency log.

TODO: add links for the GCS buckets once public.

## Claimant Model

| Role         | Description |
| -----------  | ----------- |
| **Claimant** | Transparency.dev team |
| **Claim**    | <ol><li>The digest of the bootloader is derived from this source Github repository, and is reproducible.</li><li>The bootloader firmware is issued by the Transparency.dev team.</li></ol> |
| **Believer** | The [provision](https://github.com/transparency-dev/armored-witness/tree/main/cmd/provision) and [verify](https://github.com/transparency-dev/armored-witness/tree/main/cmd/verify) tools. |
| **Verifier** | <ol><li>For Claim #1: third party auditing the Transparency.dev team</li><li>For Claim #2: the Transparency.dev team</li></ol> |
| **Arbiter**  | Log ecosystem participants and reliers |

The **Statement** is defined in
[https://github.com/transparency-dev/armored-witness-common/tree/main/release/firmware/ftlog/log_entries.go](https://github.com/transparency-dev/armored-witness-common/tree/main/release/firmware/ftlog/log_entries.go).
An example is available at
[https://github.com/transparency-dev/armored-witness-common/tree/main/release/firmware/ftlog//example_firmware_release.json](https://github.com/transparency-dev/armored-witness-common/tree/main/release/firmware/ftlog//example_firmware_release.json).