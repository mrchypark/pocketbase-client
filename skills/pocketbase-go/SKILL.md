---
name: pocketbase-go
description: Use when managing a PocketBase backend in a Go project and you need safe schema changes, schema sync, or generated Go models.
---

# PocketBase Go

Manage PocketBase schema changes through the provided tooling, then immediately sync the database state back into committed code.

## Critical Rules

1. **No JSON editing**: never hand-edit `pb_schema.json`.
2. **Tool first**: use `python3 pb_migrate_ops.py ...` for schema changes instead of editing generated artifacts.
3. **Sync immediately**: after every schema change, run `make sync`.
4. **Inspect the diff**: after `make sync`, review the generated changes before moving on.

## Project layout

- `.pb/`: local PocketBase binary and runtime data, usually gitignored
- `pb_migrations/`: committed schema migrations
- `pb_hooks/`: committed server-side hooks when the project uses them
- `models/`: generated Go code
- `pb_schema.json`: generated schema snapshot for codegen

## Workflow

1. **Bootstrap once**:
   - Resolve this skill directory and copy the files from `resources/` into the project root when the project does not already have them.
   - Initialize the environment with `make init` if the project is not set up yet.
2. **Preflight before changing schema**:
   - Confirm the target collection exists or should be created.
   - Check whether a similarly named field or relation already exists.
3. **Make one schema change at a time**:
   ```bash
   python3 pb_migrate_ops.py create_collection --name "posts" --type "base"
   python3 pb_migrate_ops.py add_field --collection "posts" --field-json '{"name": "title", "type": "text", "required": true}'
   python3 pb_migrate_ops.py delete_field --collection "posts" --field-name "old_field"
   ```
4. **Sync and inspect**:
   ```bash
   make sync
   git diff -- pb_migrations/ pb_schema.json models/
   ```
5. **Only then continue** to the next schema change or application code update.

## Commands Reference

- `make init`: set up the local environment
- `make serve`: start the local server
- `make sync`: dump schema from DB, refresh `pb_schema.json`, and regenerate Go code

## Deployment (Docker)

A `Dockerfile` is provided in resources. It packages the PocketBase binary and your local `pb_migrations`.

1.  **Build**: `docker build -t my-pocketbase .`
2.  **Run**: `docker run -p 8090:8090 -v pb_data:/pb/pb_data my-pocketbase`

## Best Practices

- **Verify first**: check that the target collection and field names are correct before changing schema.
- **One at a time**: perform one schema change, then sync and inspect the diff.
- **Relations**: confirm the target collection ID or name before adding relation fields.
- **Treat generated files as outputs**: if the generated diff looks wrong, fix the migration input instead of editing the outputs.
