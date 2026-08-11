# PocketBase Go Skill

This skill provides a robust, atomic, and type-safe workflow for managing PocketBase backends in Go projects. It automates schema migration, code generation, and environment setup.

In this repository, the skill lives under `skills/domain/pocketbase-go/`.

## Installation

### Global Toolkit Install

```bash
sh install/global-install.sh "$(pwd)"
sh install/verify-install.sh "$(pwd)"
```

### Project Template

```bash
sh install/project-bootstrap.sh "$(pwd)" /path/to/project
```

If you only want this skill in a checked-in project, copy or sync `skills/domain/pocketbase-go/` into the project's `.agents/skills/` directory.

## Usage

1.  **Load the Skill**: Tell your AI agent "Use the PocketBase Go skill".
2.  **Initialize**: The agent will run `make init` to setup the environment.
3.  **Manage Schema**: Ask the agent to "Create a posts collection" or "Add a title field to posts".
4.  **Sync**: The agent will run `make sync` to update `pb_schema.json` and generate Go code.

## Requirements

*   **Go 1.21+**: Required for `pbc-gen`.
*   **Python 3**: Required for the migration script.
*   **Make**: Required for the build system.
*   **Docker**: (Optional) For containerized deployment.

## Repository Structure

```text
skills/
└── domain/
    └── pocketbase-go/
        ├── SKILL.md
        ├── README.md
        └── resources/
            ├── Makefile
            ├── pb_migrate_ops.py
            └── Dockerfile
```
