# Documentation Index

English | [简体中文](README_ZH.md)

This directory contains project documentation for development, deployment, API assets, design notes, and community discussions around PaoPao.

If you are looking for the main project entry points, start from:

- [../README.md](../README.md) - project overview
- [../INSTALL.md](../INSTALL.md) - installation and local setup guide
- [../deploy/README.md](../deploy/README.md) - maintained Simple/Standard operations

## Documentation Sections

| Section | Description |
| --- | --- |
| [openapi/](openapi/) | Exported API documentation assets, including generated OpenAPI files and static docs artifacts |
| [proposal/](proposal/) | Product ideas, design proposals, and implementation notes |
| [deploy/](deploy/) | Unmaintained historical deployment references; do not copy their credentials or commands into a current system |
| [discuss/](discuss/) | Discussion-oriented documents, notes, and related references |

## Recommended Reading Paths

### New contributors

1. Read [../README.md](../README.md)
2. Follow [../INSTALL.md](../INSTALL.md)
3. Browse [proposal/](proposal/) to understand product intent and feature direction

### Operators and self-hosters

1. Start with [../INSTALL.md](../INSTALL.md)
2. Use [../deploy/README.md](../deploy/README.md) for maintained Simple/Standard operations
3. Consult [deploy/](deploy/) only as historical migration context and revalidate every detail

### API and integration work

1. Review [openapi/](openapi/)
2. Check [proposal/](proposal/) for feature context when endpoint behavior is tied to product design

## Notes

- `openapi/` is primarily a generated documentation asset bundle rather than a hand-written narrative section.
- `proposal/` contains historical and ongoing design material, so some documents reflect exploration rather than final behavior.
- `discuss/` may include working notes and community-oriented references in addition to formal documentation.
- Files below `docs/deploy/` can contain obsolete example credentials and are not covered by current cloud smoke tests.
