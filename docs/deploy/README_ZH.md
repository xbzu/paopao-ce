# 部署文档索引

[English](README.md) | 简体中文

本目录保留旧版及特定云平台部署资料。它们不属于当前维护的 Simple/Standard
部署流程，也没有被当前云端冒烟测试覆盖。请勿把其中的命令、镜像名、端口或配置
示例当作当前默认值。

当前请使用以下维护中的文档：

- [../../deploy/README_ZH.md](../../deploy/README_ZH.md) - Simple/Standard 日常管理
- [../../INSTALL_ZH.md](../../INSTALL_ZH.md) - 安装与运行配置指南
- [../README_ZH.md](../README_ZH.md) - 文档总索引

## 部署分区

| 分区 | 说明 |
| --- | --- |
| [core/](core/) | 旧版核心配置参考 |
| [local/](local/) | 旧版本地依赖说明 |
| [aliyun/](aliyun/) | 历史阿里云资料 |
| [huawei/](huawei/) | 历史华为云资料 |
| [tencent/](tencent/) | 历史腾讯云资料 |
| [k8s/](k8s/) | 历史 Kubernetes 资料 |

## 说明

- 这些文件只用于迁移研究和历史参考。
- 使用前必须与当前顶层文档逐项核对，不能直接照抄运行。
- Kubernetes 或云厂商专用路线需要重新设计并通过测试后，才能标记为受支持。
