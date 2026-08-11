# PocketBase Client (Go) Skill

This skill captures how to build Go applications on top of a PocketBase backend using the `github.com/mrchypark/pocketbase-client` library — from installing the client and generating typed models with `pbc-gen`, to typed CRUD, authentication, realtime, batch, files, and error handling.

In this repository, the skill lives under `skills/domain/pocketbase-client/`.

## Installation

### Global Toolkit Install

```bash
sh install/global-install.sh "$(pwd)"
sh install/verify-install.sh "$(pwd)"
```

### Project Local

If you only want this skill available in a checked-in project, copy or sync `skills/domain/pocketbase-client/` into the project's `.agents/skills/` (or equivalent) directory.

## Usage

1. **Load the Skill**: Tell your agent "Use the pocketbase-client skill".
2. **Set up**: Install the client and `pbc-gen`, export `GET /api/collections` to `schema.json`, and run `pbc-gen` to produce `models.gen.go`.
3. **Develop**: Ask the agent to build CRUD, add auth, subscribe to realtime events, upload files, or run batch operations using the typed or dynamic API.
4. **Regenerate**: After schema changes, rerun `pbc-gen` — never hand-edit the generated models.

## Requirements

- **Go 1.25+**
- A running PocketBase server for live development (`http://127.0.0.1:8090` by default)
- `pbc-gen` for code generation

## Repository Structure

```text
skills/
└── domain/
    └── pocketbase-client/
        ├── SKILL.md
        └── README.md
```
