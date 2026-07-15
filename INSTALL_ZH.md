# 安装指南

[English](INSTALL.md) | 简体中文

本文档只保留当前可验证的安装路径。项目概览见 [README_ZH.md](README_ZH.md)，日常部署命令见 [deploy/README_ZH.md](deploy/README_ZH.md)。

## 推荐路线：Docker Compose

### 环境要求

- Docker Engine 或 Docker Desktop
- Docker Compose `2.20+`；`./manage.sh up` 的等待能力需要支持 `--wait`
- Linux、macOS，或带 Docker Desktop/WSL 的 Windows

### Simple：默认轻量版

Simple 使用 SQLite 和本地文件存储。Redis 与 Meilisearch 由 Compose 自动管理，只在内部网络开放；宿主机默认仅绑定 `127.0.0.1:8008`。

```sh
git clone https://github.com/xbzu/paopao-ce.git
cd paopao-ce
./manage.sh up
```

脚本会自动：

1. 创建权限为 `0600` 的 `.env`；
2. 生成 JWT、后台加密和服务依赖所需的随机密钥；
3. 构建 Web 与 Go 一体镜像；
4. 等待依赖健康；
5. 对全新数据库执行自动迁移。

启动后访问 `http://localhost:8008`。先在页面注册站长账号，再执行：

```sh
./manage.sh promote-admin 你的用户名
```

项目不内置通用管理员账号或默认密码。

### Standard：MySQL 版

```sh
./manage.sh standard up
./manage.sh standard promote-admin 你的用户名
```

Standard 使用独立的 Compose 项目和数据卷。它不会自动继承 Simple 的 SQLite 数据或上传文件。切换前先备份并停止原模式；正式迁移需要单独的数据迁移方案。

### 日常管理

```sh
./manage.sh status
./manage.sh logs
./manage.sh logs redis
./manage.sh restart
./manage.sh check
./manage.sh down
```

Standard 命令在操作名前加 `standard`，例如 `./manage.sh standard status`。

### 域名与对外访问

默认 `PAOPAO_BIND_ADDRESS=127.0.0.1`，适合在同一台机器上通过 Nginx、Caddy 或其他反向代理提供 HTTPS。准备直接对外监听时，才把它改为 `0.0.0.0`，并同时配置防火墙、TLS 和可信反向代理。

正式域名还需要：

- 把 `.env` 中的 `PAOPAO_PUBLIC_HOST` 改成实际域名；
- HTTPS 站点同时设置 `PAOPAO_PUBLIC_SECURE=true`；
- 在管理后台确认由环境变量托管的 Local OSS 公开域名与 HTTPS 状态（托管部署中为只读）；
- 修改 `.env` 后，Simple 再执行 `./manage.sh up`，Standard 再执行 `./manage.sh standard up`，让 Compose 重建受影响容器；
- 后台标记为“重启后生效”的配置，Simple 使用 `./manage.sh restart`，Standard 使用 `./manage.sh standard restart`。

`restart` 只会重启现有应用容器，不会重新读取 `.env` 中已变化的容器环境变量。

不要直接复制 `.env.example` 的占位密钥。`manage.sh` 会拒绝占位值和过短密钥。

### 数据与备份

- `.env` 与数据必须一起备份；丢失 `PAOPAO_ADMIN_SETTINGS_KEY` 后，后台保存的敏感值无法解密。
- Standard 数据卷初始化后，只修改 `.env` 的 `PAOPAO_MYSQL_PASSWORD` 或 `PAOPAO_MYSQL_ROOT_PASSWORD` 都不会轮换已有 MySQL 账户。应先在数据库内轮换对应账户密码，再同步 `.env` 并执行 `./manage.sh standard up`；受控轮换命令留待后续实现。
- `./manage.sh down` 保留数据卷，不会删除数据。
- Simple 的 SQLite 使用 WAL；备份前停止应用，不能只在运行中复制单个 `.db` 文件。
- Standard 需要分别备份 MySQL、上传卷、Redis/Meilisearch 数据以及 `.env`。

## 源码开发

### 环境要求

- Go `1.24+`
- Node.js `20.19+` 或 `22.12+`
- Yarn `1.22.x`
- 与所选 Feature 对应的 MySQL 或 SQLite、Redis、Meilisearch

### 后端与内嵌 Web

```sh
git clone https://github.com/xbzu/paopao-ce.git
cd paopao-ce
cp config.yaml.sample config.yaml
cd web
corepack yarn install --frozen-lockfile
cd ..
make run
```

运行前编辑 `config.yaml`：源码直接运行时，把示例中的 `db`、`redis`、`meili` 等 Compose 服务名改成 `localhost` 或真实地址，并替换所有 `CHANGE-ME` 值。

全新数据库由 `Migration` 自动建表，不要先导入 `scripts/paopao-*.sql`。这些整库 SQL 仅用于旧流程兼容；与全新自动迁移混用会产生表冲突。

`make run`、`make build` 和发布目标都会先重新构建 Web 资源，避免把仓库中的旧预编译页面打进二进制。

```sh
make build
make test
```

`make release` 生成独立二进制压缩包，不包含 Docker 的源码构建上下文。需要使用
`manage.sh` 与当前维护的 Compose 部署时，请克隆完整仓库。

### 前端独立开发

```sh
cd web
corepack yarn install --frozen-lockfile
cp .env .env.local
corepack yarn dev
```

`.env` 默认使用同源 API。只有前后端分开运行时，才在 `.env.local` 设置 `VITE_HOST`。

### 桌面端

安装 Rust 和 Tauri 对应平台依赖后：

```sh
cd web
corepack yarn tauri build
```

## 配置加载顺序

应用先加载二进制内置默认值，然后：

1. 若设置 `PAOPAO_CONFIG_FILE`，加载该明确文件；
2. 否则查找 `./config.yaml`，再查找 `./custom/config.yaml`；
3. 最后用 `PAOPAO_JWT_SECRET`、`PAOPAO_ADMIN_SETTINGS_KEY`、`PAOPAO_MEILI_MASTER_KEY`、`PAOPAO_MYSQL_PASSWORD`、`PAOPAO_PUBLIC_HOST` 和 `PAOPAO_PUBLIC_SECURE` 覆盖对应部署字段。

环境变量只覆盖目标字段，不会清空同组的主机、索引或过期时间配置。

## 验证边界

GitHub Actions 会在云端分别启动 Simple 与 Standard，检查健康接口、Web 页面、注册、管理员晋升、登录和后台配置接口。构建或云端测试通过不等同于已经在本机或生产服务器部署。
