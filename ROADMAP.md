# PaoPao CE roadmap

This roadmap describes the independently maintained `xbzu/paopao-ce` line.
Historical upstream plans are preserved in the original repository history.

## Phase 1: operable baseline

- [x] Provide one management entrypoint for initialization, lifecycle, logs, and administrator promotion.
- [x] Offer separate Simple (SQLite) and Standard (MySQL) Compose modes.
- [x] Generate deployment secrets, bind to localhost by default, and run migrations automatically.
- [x] Add an administrator settings workspace with runtime overview, search, validation, and restart status.
- [x] Keep only atomically enforced registration and phone gates live; mark other mutable settings restart-required.
- [x] Test both deployment modes in GitHub Actions without publishing a production instance.

## Phase 2: administration and safety

- [ ] Promote selected restart-required settings to atomic runtime snapshots where immediate updates add real value.
- [ ] Add administrator user, content, report, and moderation workflows.
- [ ] Add configuration audit history, rollback, and permission scopes.
- [ ] Add coordinated secret rotation for MySQL, Meilisearch, JWT, and encrypted administrator settings.
- [ ] Add backup, restore, and guided Simple-to-Standard data migration commands.
- [ ] Add structured upgrade checks and release notes.

## Phase 3: product evolution

- [ ] Improve onboarding, mobile layouts, accessibility, and localization.
- [ ] Add observability presets and operator dashboards.
- [ ] Expand abuse prevention, rate limiting, and account security controls.
- [ ] Evaluate modular integrations only after their operational and security boundaries are tested.

Changes land on `main` through focused pull requests. A checked roadmap item means
the code and repository automation exist; it does not mean a production deployment
has been performed.
