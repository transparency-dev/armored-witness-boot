# Recovery

The Dockerfile in this directory is used for building an image to be used as a
recovery/provisioning tool for the Armored Witness.

We currently use the firmware in the github.com/usbarmory/armory-ums repo as
the recovery tool.

While that repo does offer prebuilt binary releases, we rebuild from scratch
here so we can be sure about which TamaGo toolchain version is used, etc.

## Build and Release Process

A
[Cloud Build trigger](https://cloud.google.com/build/docs/automating-builds/create-manage-triggers)
is defined by a yaml config file. The Transparency.dev team invokes it manually
when we want to publish a release.

The pipeline includes two main steps: building and making available the recovery
tool files, and writing the release metadata (Claimant Model Statement) to the
firmware transparency log.

1.  Cloud Build builds the recovery builder Docker image and copies the compiled
    recovery imx file to a public Google Cloud Storage bucket.
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
| **Claim**    | <ol><li>The digest of the recovery tool is derived from this source Github repository, and is reproducible.</li><li>The recovery tool is issued by the Transparency.dev team.</li></ol> |
| **Believer** | The [provision](https://github.com/transparency-dev/armored-witness/tree/main/cmd/provision) and [verify](https://github.com/transparency-dev/armored-witness/tree/main/cmd/verify) tools. |
| **Verifier** | <ol><li>For Claim #1: third party auditing the Transparency.dev team</li><li>For Claim #2: the Transparency.dev team</li></ol> |
| **Arbiter**  | Log ecosystem participants and reliers |

The **Statement** is defined in
[https://github.com/transparency-dev/armored-witness-common/tree/main/release/firmware/ftlog/log_entries.go](https://github.com/transparency-dev/armored-witness-common/tree/main/release/firmware/ftlog/log_entries.go).
An example is available at
[https://github.com/transparency-dev/armored-witness-common/tree/main/release/firmware/ftlog//example_firmware_release.json](https://github.com/transparency-dev/armored-witness-common/tree/main/release/firmware/ftlog//example_firmware_release.json).