<div id="top"></div>

[![Go](https://github.com/xbzu/paopao-ce/actions/workflows/go.yml/badge.svg)](https://github.com/xbzu/paopao-ce/actions/workflows/go.yml)
[![Deployment smoke test](https://github.com/xbzu/paopao-ce/actions/workflows/deploy.yml/badge.svg)](https://github.com/xbzu/paopao-ce/actions/workflows/deploy.yml)
[![Stars](https://img.shields.io/github/stars/xbzu/paopao-ce.svg?style=flat)](https://github.com/xbzu/paopao-ce/stargazers)
[![MIT License](https://img.shields.io/github/license/xbzu/paopao-ce.svg?style=flat)](https://github.com/xbzu/paopao-ce/blob/main/LICENSE)

<div align="center">
  <a href="https://github.com/xbzu/paopao-ce">
    <img src="./.assets/readme/paopao-logo.png" alt="PaoPao logo" width="88" height="88">
  </a>

  <h1 align="center">PaoPao</h1>

  <p align="center">
    一个基于 Go 与 Vue 的开源微社区平台。
    <br />
    适合自部署社区产品、功能验证，以及在现有能力之上进行二次开发。
  </p>

  <p align="center">
    <a href="README.md">English</a>
    ·
    <a href="https://github.com/xbzu/paopao-ce/pulls">Pull Requests</a>
    ·
    <a href="INSTALL_ZH.md">安装说明</a>
  </p>
</div>

---

## 项目简介

PaoPao 是一个完整的开源微社区系统，包含 Go 后端、Vue 3 Web 前端，以及可选的 Tauri 桌面端。项目通过模块化的特性开关体系，将存储、搜索、日志、可观测性与部署方式组合在一起，便于按需裁剪和扩展。

本仓库是独立维护的二次开发版本，不保留 GitHub Fork 关系；原项目版权与 MIT License 继续完整保留。

第一阶段暂时保留 Go module 路径 `github.com/rocboss/paopao-ce`，以避免一次性改动全部 import、生成代码和外部兼容接口；这只是代码兼容标识，不代表保留 GitHub 上游关系。

如果你希望快速搭建一个可运行的社区产品，或者基于成熟代码库继续定制开发，而不是从零开始，PaoPao 是一个很合适的起点。

## 为什么选择 PaoPao

- **全栈一体化**：后端、Web 前端、桌面端构建都在同一仓库中维护。
- **模块化运行能力**：GitHub CI 同时覆盖 `Simple` 与 `Standard` 两套能力组合。
- **基础设施灵活**：支持 MySQL、PostgreSQL、SQLite、Redis、Meilisearch 以及多种对象存储后端。
- **部署入口统一**：通过一个管理脚本完成密钥初始化、启动、状态、日志与管理员设置。
- **适合自部署**：配置文件清晰，仓库内已提供部署与开发文档。

## 预览

### Web 端

![明色主题预览](./.assets/readme/preview-light.jpeg)

![暗色主题预览](./.assets/readme/preview-dark.jpeg)

### 桌面端

![](docs/proposal/.assets/000-00.jpg)

## 架构概览

| 层级 | 主要技术 |
| --- | --- |
| 后端 | Go、Gin、Cobra、GORM、Mir |
| Web 前端 | Vue 3、Vite、Naive UI |
| 桌面端 | Tauri |
| 搜索 | Meilisearch |
| 缓存 | Redis |
| 对象存储 | Local OSS、MinIO、AliOSS、COS、Huawei OBS、S3 兼容存储 |
| 可观测性 | OpenTelemetry、Sentry、Pyroscope、Pprof |

## 仓库结构

| 路径 | 说明 |
| --- | --- |
| `cmd/`、`internal/`、`pkg/` | 后端应用与共享包 |
| `web/` | Vue 3 Web 应用与 Tauri 桌面端前端 |
| `docs/` | 部署文档、OpenAPI 文档、设计提案等 |
| `scripts/` | SQL 初始化脚本与辅助资源 |
| `config.yaml.sample` | 最小化启动配置模板 |

## 快速开始

### 方案 A：一条命令启动轻量版

轻量版使用 SQLite 与本地文件存储；Redis 和 Meilisearch 由 Compose 自动管理且不暴露宿主机端口。首次启动会自动生成随机密钥并执行数据库迁移。

```sh
git clone https://github.com/xbzu/paopao-ce.git
cd paopao-ce
./manage.sh up
```

启动后访问 `http://localhost:8008`。注册站长账号后，将它提升为管理员：

```sh
./manage.sh promote-admin 你的用户名
```

常用命令为 `./manage.sh status`、`./manage.sh logs`、`./manage.sh restart` 和 `./manage.sh down`。请在首次启动前选定 Simple 或 Standard；后续切换模式需先备份、执行 `down`，并单独迁移数据。完整说明见 [部署与管理指南](deploy/README_ZH.md)。

### 方案 B：从源码开发

源码开发需要 Go `1.24+`、Node.js `20.19+` 或 `22.12+`、Yarn `1.22.x`，以及所选 Feature 对应的数据库、Redis 和 Meilisearch。

先安装前端依赖，再复制并修改启动配置。源码直接运行时，需要把 `db`、`redis`、`meili` 等 Compose 服务名改成 `localhost` 或实际地址，并替换所有 `CHANGE-ME` 值。

```sh
cd web
corepack yarn install --frozen-lockfile
cd ..
cp config.yaml.sample config.yaml
make run
```

全新数据库由 Migration 自动建表，不要先导入 `scripts/paopao-*.sql`。`make run`、`make build` 和发布目标都会先重建并内嵌 Web 页面。

```sh
make build
make test
```

前后端分开调试、桌面端构建、迁移和配置优先级说明见 [INSTALL_ZH.md](INSTALL_ZH.md)。

## 配置与 Feature 套件

`config.yaml.sample` 现在是一个**最小化 bootstrap 模板**。运行时，PaoPao 会先加载内置默认配置；若设置了 `PAOPAO_CONFIG_FILE`，再叠加该明确指定的文件，否则依次查找 `./config.yaml` 与 `./custom/config.yaml`。Compose 固定使用 `/app/config.yaml`，不再依赖模糊的搜索顺序。

这意味着外部配置文件可以有意保持精简。当前建议是：

- **继续放在 YAML 中**：端口、Feature 组合、数据库、Redis、JWT，以及 `AdminSettings.EncryptionKey`
- **尽量改到管理后台**（`/#/admin/settings`）：站点资料、应用行为阈值、全文索引、对象存储、短信/支付等已接入 registry 的配置
- **注意生效方式**：部分配置即时生效，部分会在后台中明确标记为“重启后生效”

`Features` 配置用于控制启用哪些能力组合：

```yaml
Features:
  Default: ["Standard"]
  Simple: ["Web", "Frontend:EmbedWeb", "Sqlite3", "LocalOSS", "Redis", "Meili", "Migration"]
  Standard: ["Web", "Frontend:EmbedWeb", "MySQL", "LocalOSS", "Redis", "Meili", "Migration"]
```

常见用法示例：

```sh
# 使用默认套件
release/paopao serve

# 仅使用轻量套件
release/paopao serve --no-default-features --features simple

# 手动显式指定功能项
release/paopao serve --no-default-features --features web,frontend:embedweb,sqlite3,localoss,redis,meili,migration
```

[features-status.md](features-status.md) 保留的是历史 Feature 说明。当前部署契约
以以上 Simple/Standard 套件及其 GitHub 冒烟测试为准。

## 文档导航

- [INSTALL_ZH.md](INSTALL_ZH.md) - 中文安装、Docker 使用与桌面端构建说明
- [INSTALL.md](INSTALL.md) - English installation guide
- [docs/README_ZH.md](docs/README_ZH.md) - 中文文档总索引
- [docs/README.md](docs/README.md) - English documentation index
- [docs/deploy/](docs/deploy/) - 部署相关文档
- [docs/openapi/](docs/openapi/) - 导出的 API 文档资源
- [docs/proposal/](docs/proposal/) - 设计说明与功能提案
- [deployed-sites.md](deployed-sites.md) - 已知部署站点列表
- [ROADMAP.md](ROADMAP.md) - 路线图与规划说明

## 分支策略

本仓库采用精简分支模型：

| 分支 | 作用 |
| --- | --- |
| `main` | 已通过自动测试的稳定基线 |
| `agent/*` | 自动化改进与验收分支 |
| `feature/*` | 人工开发的独立功能分支 |

所有改动通过 Pull Request 合并到 **`main`**。

## 参与贡献

欢迎提交 Pull Request。建议的贡献流程：

1. Fork 或克隆当前仓库。
2. 从 `main` 拉出自己的功能分支。
3. 保持改动聚焦，并补充必要文档。
4. 提交 PR 时说明问题背景、实现方式与验证结果。

如果你已经部署了自己的实例，也欢迎将站点补充到 [deployed-sites.md](deployed-sites.md)。

## License

项目基于 MIT License 发布，详见 [LICENSE](LICENSE)。

本二次开发仓库保留原作者版权声明。原项目与历史项目笔记可分别在
[rocboss/paopao-ce](https://github.com/rocboss/paopao-ce) 和
[语雀项目笔记](https://www.yuque.com/rocs/paopao/about) 查看。

<p align="right">(<a href="#top">回到顶部</a>)</p>
