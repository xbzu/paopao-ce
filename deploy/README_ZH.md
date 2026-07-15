# PaoPao 部署与管理指南

## 选择部署模式

| 模式 | 适合场景 | 数据库 | 对外端口 |
| --- | --- | --- | --- |
| Simple（默认） | 单机、自用、小型社区、功能验证 | SQLite | 仅 8008 |
| Standard | 数据量增长、需要独立数据库 | MySQL | 仅 8008 |

两种模式都会自动管理 Redis 与 Meilisearch。它们只在 Compose 内部网络提供服务，不会把数据库、缓存或搜索端口暴露到宿主机。

## 首次启动

环境要求：Docker Engine 或 Docker Desktop，Docker Compose `2.20+`，并支持 `--wait`。

```sh
./manage.sh up
```

这条命令会：

1. 创建权限为 `0600` 的 `.env`；
2. 生成 JWT、后台敏感配置和 Meilisearch 随机密钥；
3. 构建前后端一体镜像；
4. 等待 Redis 与 Meilisearch 就绪；
5. 自动把全新 SQLite 数据库迁移到最新版本；
6. 只开放 `http://localhost:8008`。

首次启动后，在网页注册站长账号，再执行：

```sh
./manage.sh promote-admin 你的用户名
```

项目不会预置通用管理员或默认密码。提升成功后，可从侧栏进入“管理后台”。

## 标准版

```sh
./manage.sh standard up
./manage.sh standard promote-admin 你的用户名
```

Standard 模式增加 MySQL，并继续使用自动迁移，不再混用旧的整库初始化 SQL。

## 日常管理

```sh
./manage.sh status          # 查看容器状态
./manage.sh logs            # 跟踪应用日志
./manage.sh logs redis      # 查看指定服务日志
./manage.sh restart         # 仅重启应用，使“待重启”配置生效
./manage.sh down            # 停止服务，保留数据卷
./manage.sh check           # 只校验部署配置，不启动服务
```

Standard 模式在命令前加 `standard`。

修改 `.env` 或 Compose 配置后，Simple 请再次执行 `./manage.sh up`，Standard 请执行
`./manage.sh standard up`，让 Compose 重建受影响容器并应用新环境变量；`restart`
只适合后台标记为“待重启”的配置。
在这两套托管部署中，Meilisearch 主密钥只由 `.env` 的
`PAOPAO_MEILI_MASTER_KEY` 管理。

已初始化的 MySQL 数据卷不会因为 `.env` 改动自动轮换 `paopao` 用户或 root 密码。
需要轮换时，先在 MySQL 内修改对应账户，再同步 `PAOPAO_MYSQL_PASSWORD` 或
`PAOPAO_MYSQL_ROOT_PASSWORD` 并执行 `./manage.sh standard up`。

## 域名、数据与密钥

- 正式上线前，把 `.env` 中的 `PAOPAO_PUBLIC_HOST` 改成真实域名；HTTPS 站点同时设置 `PAOPAO_PUBLIC_SECURE=true`。托管部署的后台只读展示这两个值，以 `.env` 为唯一权威来源。
- `.env` 不会提交到 Git；请和数据一起安全备份。
- `PAOPAO_ADMIN_SETTINGS_KEY` 用于解密后台保存的密钥。丢失后，旧的敏感配置无法恢复。
- Simple 模式数据位于 Compose 命名卷；复制 SQLite 文件前应先停止应用，避免遗漏 WAL 内容。
- 项目不会把密钥、主机地址或文件路径返回到公开接口。

## 管理后台

`/#/admin/settings` 提供：

- 当前部署模式、版本、数据库、缓存、搜索与存储概览；
- 按中文名称、说明或配置 Key 搜索；
- 即时生效、重启后生效和未激活状态提示；
- 站点、内容、存储、搜索、短信和支付配置；
- Secret 只显示“已配置/未配置”，不会回显明文。

“允许用户注册”由后端运行策略强制执行，保存后即时生效。“允许绑定手机号”只有在启用 Sms 功能并配置短信服务后才可开放；未启用 Sms 时后端会拒绝验证码和绑定请求。当前概览展示的是配置与运行信息，不等同于对第三方服务做完整健康探测。
