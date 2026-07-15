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
    An open-source micro-community platform built with Go and Vue.
    <br />
    Designed for self-hosted social products, community experiments, and customizable deployments.
  </p>

  <p align="center">
    <a href="README_ZH.md">简体中文</a>
    ·
    <a href="https://github.com/xbzu/paopao-ce/pulls">Pull Requests</a>
    ·
    <a href="INSTALL.md">Install</a>
  </p>
</div>

---

## Overview

PaoPao is a full-stack, open-source micro-community system. It combines a Go backend, a Vue 3 web client, and an optional Tauri desktop application, with a modular feature system for storage, search, logging, observability, and deployment strategy.

This is an independently maintained derivative repository with no GitHub fork relationship. The original copyright notices and MIT License remain intact.

Phase one intentionally retains the Go module path `github.com/rocboss/paopao-ce` to avoid a disruptive import/generated-code migration. It is a compatibility identifier, not a GitHub fork relationship.

The repository is suitable for teams or individuals who want to run a community product, evaluate an extensible social platform, or build on top of an existing codebase instead of starting from scratch.

## Why PaoPao

- **Full-stack delivery**: backend, web frontend, and desktop packaging live in one repository.
- **Modular runtime features**: CI covers both the `Simple` and `Standard` capability suites.
- **Flexible infrastructure**: supports MySQL, PostgreSQL, SQLite, Redis, Meilisearch, and multiple object storage providers.
- **One deployment entrypoint**: initialize secrets, start services, inspect status and logs, and bootstrap the administrator through one management script.
- **Self-hosting friendly**: configuration is file-based and operational docs are already included in the repo.

## Preview

### Web

![Light theme preview](./.assets/readme/preview-light.jpeg)

![Dark theme preview](./.assets/readme/preview-dark.jpeg)

### Desktop

![](docs/proposal/.assets/000-00.jpg)

## Architecture at a Glance

| Layer | Primary stack |
| --- | --- |
| Backend | Go, Gin, Cobra, GORM, Mir |
| Web frontend | Vue 3, Vite, Naive UI |
| Desktop client | Tauri |
| Search | Meilisearch |
| Cache | Redis |
| Object storage | Local OSS, MinIO, AliOSS, COS, Huawei OBS, S3-compatible |
| Observability | OpenTelemetry, Sentry, Pyroscope, Pprof |

## Repository Layout

| Path | Purpose |
| --- | --- |
| `cmd/`, `internal/`, `pkg/` | Backend application and shared packages |
| `web/` | Vue 3 web application and Tauri desktop frontend |
| `docs/` | Deployment, OpenAPI, design proposals, and related documentation |
| `scripts/` | SQL bootstrap and helper assets |
| `config.yaml.sample` | Minimal bootstrap configuration template |

## Quick Start

### Option A: Start the Simple deployment with one command

Simple mode uses SQLite and local file storage. Compose manages Redis and Meilisearch internally without publishing their ports. First startup creates random secrets and applies database migrations.

```sh
git clone https://github.com/xbzu/paopao-ce.git
cd paopao-ce
./manage.sh up
```

Open `http://localhost:8008`. After registering the owner account, promote it once:

```sh
./manage.sh promote-admin YOUR_USERNAME
```

Use `./manage.sh status`, `./manage.sh logs`, `./manage.sh restart`, and `./manage.sh down` for routine operations. Choose Simple or Standard before the first startup; switching modes later requires a backup, `down`, and a separate data migration. See [Deployment and management](deploy/README.md).

### Option B: Develop from source

Source development requires Go `1.24+`, Node.js `20.19+` or `22.12+`, Yarn `1.22.x`, and the database, Redis, and Meilisearch required by the selected features.

Install Web dependencies, then copy and edit the bootstrap config. For a process running directly on the host, replace Compose service names such as `db`, `redis`, and `meili` with `localhost` or actual addresses, and replace every `CHANGE-ME` value.

```sh
cd web
corepack yarn install --frozen-lockfile
cd ..
cp config.yaml.sample config.yaml
make run
```

Migration creates a fresh schema, so do not import `scripts/paopao-*.sql` first. `make run`, `make build`, and release targets rebuild and embed the Web UI automatically.

```sh
make build
make test
```

See [INSTALL.md](INSTALL.md) for separate frontend development, desktop builds, migration behavior, and configuration precedence.

## Configuration and Feature Suites

`config.yaml.sample` is now a **minimal bootstrap template**. PaoPao first loads embedded defaults. When `PAOPAO_CONFIG_FILE` is set it overlays that exact file; otherwise it searches `./config.yaml` and then `./custom/config.yaml`. Compose always selects `/app/config.yaml` explicitly.

That means the external config can stay intentionally small. In the current design:

- **Keep in YAML**: ports, feature selection, database, Redis, JWT, and `AdminSettings.EncryptionKey`
- **Prefer the admin UI** (`/#/admin/settings`): site profile, app behavior limits, search provider settings, object-storage settings, SMS/payment settings, and other operational knobs supported by the registry
- **Watch apply mode**: some settings are live, while others are marked restart-required in the admin page

The `Features` section controls which capability bundles are enabled:

```yaml
Features:
  Default: ["Standard"]
  Simple: ["Web", "Frontend:EmbedWeb", "Sqlite3", "LocalOSS", "Redis", "Meili", "Migration"]
  Standard: ["Web", "Frontend:EmbedWeb", "MySQL", "LocalOSS", "Redis", "Meili", "Migration"]
```

Typical examples:

```sh
# Use the default suite
release/paopao serve

# Use only the simple suite
release/paopao serve --no-default-features --features simple

# Enable features explicitly
release/paopao serve --no-default-features --features web,frontend:embedweb,sqlite3,localoss,redis,meili,migration
```

For historical feature notes, see [features-status.md](features-status.md). The
maintained deployment contract is defined by the Simple/Standard suites above and
their GitHub smoke tests.

## Documentation Map

- [INSTALL.md](INSTALL.md) - installation, Docker usage, and desktop build instructions
- [INSTALL_ZH.md](INSTALL_ZH.md) - Chinese installation guide
- [docs/README.md](docs/README.md) - documentation index
- [docs/README_ZH.md](docs/README_ZH.md) - Chinese documentation index
- [docs/deploy/](docs/deploy/) - deployment documentation
- [docs/openapi/](docs/openapi/) - exported API documentation assets
- [docs/proposal/](docs/proposal/) - design notes and feature proposals
- [deployed-sites.md](deployed-sites.md) - known deployed instances
- [ROADMAP.md](ROADMAP.md) - roadmap and planning notes

## Branch Strategy

This repository uses a compact branch model:

| Branch | Role |
| --- | --- |
| `main` | Stable baseline that has passed automation |
| `agent/*` | Automated improvement and verification branches |
| `feature/*` | Focused human development branches |

All changes merge into **`main`** through a pull request.

## Contributing

Pull requests are welcome. If you want to contribute:

1. Fork or clone the repository.
2. Create a feature branch from `main`.
3. Keep changes focused and documented.
4. Open a PR with context about the problem, approach, and verification.

If you deploy your own instance, consider adding it to [deployed-sites.md](deployed-sites.md).

## License

Distributed under the MIT License. See [LICENSE](LICENSE) for details.

This derivative keeps the original copyright notices. The original project and
historical notes are available at [rocboss/paopao-ce](https://github.com/rocboss/paopao-ce)
and [its project notes](https://www.yuque.com/rocs/paopao/about).

<p align="right">(<a href="#top">back to top</a>)</p>
