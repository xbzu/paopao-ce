# Deployment Documentation

English | [简体中文](README_ZH.md)

This directory contains legacy, platform-specific deployment references. They
are not maintained by the current Simple/Standard deployment workflow and are
not covered by its cloud smoke tests. Do not treat their commands, image names,
ports, or configuration examples as current defaults.

Use the maintained documentation instead:

- [../../deploy/README.md](../../deploy/README.md) - Simple/Standard operations
- [../../INSTALL.md](../../INSTALL.md) - installation and runtime setup guide
- [../README.md](../README.md) - overall documentation index

## Deployment Sections

| Section | Description |
| --- | --- |
| [core/](core/) | Legacy core configuration references |
| [local/](local/) | Legacy local dependency setup notes |
| [aliyun/](aliyun/) | Historical Alibaba Cloud notes |
| [huawei/](huawei/) | Historical Huawei Cloud notes |
| [tencent/](tencent/) | Historical Tencent Cloud notes |
| [k8s/](k8s/) | Historical Kubernetes references |

## Notes

- These files are retained only for migration research and historical context.
- Revalidate every command against the current top-level documentation before use.
- A future Kubernetes or provider-specific route must be redesigned and tested before it can be marked supported.
