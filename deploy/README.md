# PaoPao deployment and management

Requirements: Docker Engine or Docker Desktop with Docker Compose `2.20+` and
support for `--wait`.

Use `./manage.sh up` for the default Simple deployment (SQLite), or `./manage.sh standard up` for MySQL. Both modes keep Redis, Meilisearch, and the database on an internal Compose network and publish only port 8008.

The generated configuration binds to `127.0.0.1` by default. For a production hostname, set `PAOPAO_PUBLIC_HOST`; also set `PAOPAO_PUBLIC_SECURE=true` when the public site uses HTTPS.
These two public-address values are environment-owned and read-only in the
administrator page for managed deployments.

On first startup, the script creates `.env` with random JWT, admin-settings encryption, Meilisearch, and MySQL secrets. Back up this file with your data. Register the owner account in the web UI, then run:

```sh
./manage.sh promote-admin YOUR_USERNAME
```

Routine commands are `status`, `logs`, `restart`, `down`, and `check`. Prefix them with `standard` for the MySQL deployment. The project intentionally ships no default administrator or shared password.

Use `restart` for administrator settings marked restart-required. After changing
`.env` or Compose configuration, run `up` again so Compose recreates affected
containers and applies the new environment. In the managed deployments, the
Meilisearch master key is controlled only by `PAOPAO_MEILI_MASTER_KEY` in `.env`.

An initialized MySQL volume does not automatically rotate either the `paopao`
user or root password when `.env` changes. Rotate the matching account in MySQL
first, then update `PAOPAO_MYSQL_PASSWORD` or `PAOPAO_MYSQL_ROOT_PASSWORD` and
run `./manage.sh standard up`.

The administrator page is available at `/#/admin/settings`. Secrets are never echoed in plaintext, and settings clearly distinguish live changes from restart-required changes.
