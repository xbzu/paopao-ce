# Installation Guide

English | [简体中文](INSTALL_ZH.md)

This guide contains only the currently supported installation paths. See [README.md](README.md) for the project overview and [deploy/README.md](deploy/README.md) for the short operations reference.

## Recommended path: Docker Compose

### Requirements

- Docker Engine or Docker Desktop
- Docker Compose `2.20+`; `./manage.sh up` also requires support for `--wait`
- Linux, macOS, or Windows with Docker Desktop/WSL

### Simple: the default lightweight deployment

Simple uses SQLite and local file storage. Compose manages Redis and Meilisearch on an internal network, while the host binds only `127.0.0.1:8008` by default.

```sh
git clone https://github.com/xbzu/paopao-ce.git
cd paopao-ce
./manage.sh up
```

The script automatically:

1. creates `.env` with mode `0600`;
2. generates random JWT, administrator-settings, and service secrets;
3. builds the combined Web and Go image;
4. waits for dependencies to become healthy;
5. migrates a fresh database.

Open `http://localhost:8008`, register the owner account, then run:

```sh
./manage.sh promote-admin YOUR_USERNAME
```

The project ships no shared administrator account or default password.

### Standard: MySQL deployment

```sh
./manage.sh standard up
./manage.sh standard promote-admin YOUR_USERNAME
```

Standard uses a separate Compose project and separate volumes. It does not inherit Simple's SQLite data or uploads. Back up and stop the previous mode first; moving production data requires an explicit migration plan.

### Routine management

```sh
./manage.sh status
./manage.sh logs
./manage.sh logs redis
./manage.sh restart
./manage.sh check
./manage.sh down
```

Prefix Standard commands with `standard`, for example `./manage.sh standard status`.

### Domains and public access

The default `PAOPAO_BIND_ADDRESS=127.0.0.1` is suitable for HTTPS through Nginx, Caddy, or another reverse proxy on the same machine. Change it to `0.0.0.0` only when direct public binding is intentional, with firewall, TLS, and trusted-proxy settings in place.

For a production hostname:

- set `PAOPAO_PUBLIC_HOST` in `.env`;
- set `PAOPAO_PUBLIC_SECURE=true` for HTTPS sites;
- confirm the environment-managed Local OSS public domain and HTTPS status in the administrator page (they are read-only in managed deployments);
- run `./manage.sh up` (Simple) or `./manage.sh standard up` (Standard) again after editing `.env`, so Compose recreates affected containers;
- use `./manage.sh restart` (Simple) or `./manage.sh standard restart` (Standard) for administrator settings marked as restart-required.

`restart` only restarts the existing application container. It does not reload
changed container environment variables from `.env`.

Do not copy placeholder secrets from `.env.example` into a running system. `manage.sh` rejects placeholders and short secrets.

### Data and backups

- Back up `.env` with the data. Losing `PAOPAO_ADMIN_SETTINGS_KEY` makes stored secret settings undecryptable.
- On an initialized Standard volume, changing `PAOPAO_MYSQL_PASSWORD` or `PAOPAO_MYSQL_ROOT_PASSWORD` only in `.env` does not rotate either existing MySQL account. Rotate the matching account inside MySQL first, then update `.env` and run `./manage.sh standard up`; a future managed rotation command is tracked separately.
- `./manage.sh down` preserves volumes.
- SQLite uses WAL. Stop the app before a file-level Simple backup instead of copying only the live `.db` file.
- Standard backups must cover MySQL, uploads, Redis/Meilisearch data, and `.env`.

## Source development

### Requirements

- Go `1.24+`
- Node.js `20.19+` or `22.12+`
- Yarn `1.22.x`
- MySQL or SQLite, Redis, and Meilisearch for the selected features

### Backend with embedded Web UI

```sh
git clone https://github.com/xbzu/paopao-ce.git
cd paopao-ce
cp config.yaml.sample config.yaml
cd web
corepack yarn install --frozen-lockfile
cd ..
make run
```

Edit `config.yaml` first. For a process running directly on the host, replace Compose service names such as `db`, `redis`, and `meili` with `localhost` or actual addresses, and replace every `CHANGE-ME` value.

The `Migration` feature creates a fresh schema. Do not import `scripts/paopao-*.sql` first: those full-schema files exist for legacy workflows, and mixing them with a new migration history causes table conflicts.

`make run`, `make build`, and release targets rebuild the Web UI before compiling Go, so a stale prebuilt page is not embedded.

```sh
make build
make test
```

`make release` creates standalone binary archives. Those archives intentionally
do not contain the Docker source-build context; clone the repository when using
`manage.sh` and the maintained Compose deployment.

### Separate frontend development

```sh
cd web
corepack yarn install --frozen-lockfile
cp .env .env.local
corepack yarn dev
```

The tracked `.env` uses same-origin API requests. Set `VITE_HOST` only in `.env.local` when the frontend and backend run separately.

### Desktop app

After installing Rust and the platform-specific Tauri prerequisites:

```sh
cd web
corepack yarn tauri build
```

## Configuration precedence

The application loads embedded defaults, then:

1. reads the explicit `PAOPAO_CONFIG_FILE` when set;
2. otherwise checks `./config.yaml` and then `./custom/config.yaml`;
3. finally applies `PAOPAO_JWT_SECRET`, `PAOPAO_ADMIN_SETTINGS_KEY`, `PAOPAO_MEILI_MASTER_KEY`, `PAOPAO_MYSQL_PASSWORD`, `PAOPAO_PUBLIC_HOST`, and `PAOPAO_PUBLIC_SECURE` to their deployment fields.

Environment overrides change only their target fields and preserve sibling host, index, and expiration settings.

## Verification boundary

GitHub Actions starts both Simple and Standard in the cloud and checks health, the embedded Web UI, registration, administrator promotion, login, and the settings API. A successful build or cloud smoke test does not mean that the project has been deployed locally or to production.
