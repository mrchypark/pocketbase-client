# PocketBase Schema Format (Legacy vs Latest)

This document captures **schema JSON structure and field options** needed for a code generator. It compares the legacy format (v0.22.x and earlier) with the latest format (v0.23+), and lists field-level option keys based on the PocketBase codebase and official docs.

## Version Summary

| Area | Legacy (v0.22.x and earlier) | Latest (v0.23+) |
| --- | --- | --- |
| Collection fields key | `schema` | `fields` |
| Field options placement | Nested under `options` | Flattened at field top-level |
| Field modeling | Single `SchemaField` + generic `Options` | Field-type specific structs (e.g., `TextField`, `RelationField`) |
| Collection options | Generic `options` map | Embedded, structured option blocks |
| System fields | Implicit | Explicit fields (e.g., `id`, `created`, `updated`) |

## Latest Format (v0.23+)

### Collection Object (Top-Level Fields)
Based on the official API docs for collections and current PocketBase models:

- `id` (string)
- `name` (string)
- `type` (string: `base`, `auth`, `view`)
- `system` (boolean)
- `fields` (array of field objects)
- `indexes` (array of SQL index definitions)
- `listRule`, `viewRule`, `createRule`, `updateRule`, `deleteRule` (string or null)

References:
- https://pocketbase.io/docs/api-collections/
- https://github.com/pocketbase/pocketbase/blob/master/core/collection_model.go

### Field Objects — Common Keys
All fields share the following properties:

- `id` (string)
- `name` (string)
- `type` (string)
- `system` (boolean)
- `hidden` (boolean)
- `presentable` (boolean)

Reference:
- https://github.com/pocketbase/pocketbase/blob/master/core/fields_list.go

### Field Objects — Type-Specific Keys
The following keys are **top-level** in v0.23+ (flattened; no `options` object). These come directly from the field structs in `core/field_*.go`.

| Field Type | Source | Keys |
| --- | --- | --- |
| `text` | https://github.com/pocketbase/pocketbase/blob/master/core/field_text.go | `min`, `max`, `pattern`, `autogeneratePattern`, `required`, `primaryKey` |
| `number` | https://github.com/pocketbase/pocketbase/blob/master/core/field_number.go | `min`, `max`, `onlyInt`, `required` |
| `bool` | https://github.com/pocketbase/pocketbase/blob/master/core/field_bool.go | `required` |
| `date` | https://github.com/pocketbase/pocketbase/blob/master/core/field_date.go | `min`, `max`, `required` |
| `autodate` | https://github.com/pocketbase/pocketbase/blob/master/core/field_autodate.go | `onCreate`, `onUpdate` |
| `select` | https://github.com/pocketbase/pocketbase/blob/master/core/field_select.go | `values`, `maxSelect`, `required` |
| `relation` | https://github.com/pocketbase/pocketbase/blob/master/core/field_relation.go | `collectionId`, `cascadeDelete`, `minSelect`, `maxSelect`, `required` |
| `file` | https://github.com/pocketbase/pocketbase/blob/master/core/field_file.go | `maxSize`, `maxSelect`, `mimeTypes`, `thumbs`, `protected`, `required` |
| `json` | https://github.com/pocketbase/pocketbase/blob/master/core/field_json.go | `maxSize`, `required` |
| `email` | https://github.com/pocketbase/pocketbase/blob/master/core/field_email.go | `exceptDomains`, `onlyDomains`, `required` |
| `url` | https://github.com/pocketbase/pocketbase/blob/master/core/field_url.go | `exceptDomains`, `onlyDomains`, `required` |
| `editor` | https://github.com/pocketbase/pocketbase/blob/master/core/field_editor.go | `maxSize`, `convertURLs`, `required` |
| `password` | https://github.com/pocketbase/pocketbase/blob/master/core/field_password.go | `pattern`, `min`, `max`, `cost`, `required` |

### Latest JSON Example
From API docs (`fields` array with flattened keys):

```json
{
  "id": "_pbc_2287844090",
  "name": "posts",
  "type": "base",
  "system": false,
  "fields": [
    {
      "id": "text3208210256",
      "name": "id",
      "type": "text",
      "system": true,
      "required": true,
      "primaryKey": true,
      "autogeneratePattern": "[a-z0-9]{15}"
    },
    {
      "id": "text724990059",
      "name": "title",
      "type": "text",
      "min": 3,
      "max": 100,
      "required": true
    },
    {
      "id": "relation123456789",
      "name": "author",
      "type": "relation",
      "collectionId": "_pbc_344172009",
      "maxSelect": 1,
      "cascadeDelete": false
    },
    {
      "id": "autodate2990389176",
      "name": "created",
      "type": "autodate",
      "system": true,
      "onCreate": true,
      "onUpdate": false
    },
    {
      "id": "autodate3332085495",
      "name": "updated",
      "type": "autodate",
      "system": true,
      "onCreate": true,
      "onUpdate": true
    }
  ],
  "indexes": [],
  "listRule": null,
  "viewRule": null,
  "createRule": null,
  "updateRule": null,
  "deleteRule": null
}
```

Reference:
- https://pocketbase.io/docs/api-collections/

## Legacy Format (v0.22.x and earlier)

### Collection Object
- `schema` (array of field objects)
- `options` (map) — collection-level options (e.g., auth/view collection settings)

References:
- https://github.com/pocketbase/pocketbase/blob/v0.22.22/models/collection.go
- https://pocketbase.io/old/docs/api-collections

### Field Object (Legacy)
Legacy fields use a **generic** `SchemaField` with a nested `options` object.

References:
- https://github.com/pocketbase/pocketbase/blob/v0.22.22/models/schema/schema_field.go
- https://pocketbase.io/old/docs/go-collections
- https://pocketbase.io/old/docs/js-collections

### Legacy JSON Example

```json
{
  "name": "posts",
  "type": "base",
  "schema": [
    {
      "id": "f1id",
      "name": "title",
      "type": "text",
      "required": true,
      "options": {
        "min": 5,
        "max": 100,
        "pattern": ""
      }
    },
    {
      "id": "f2id",
      "name": "author",
      "type": "relation",
      "options": {
        "collectionId": "ae1234567890abc",
        "cascadeDelete": false,
        "minSelect": null,
        "maxSelect": 1
      }
    }
  ],
  "options": {}
}
```

### Legacy Field Options (Nested in `options`)
From old docs and v0.22 schema types:

- `text`: `min`, `max`, `pattern`
- `number`: `min`, `max`, `noDecimal`
- `relation`: `collectionId`, `cascadeDelete`, `minSelect`, `maxSelect`
- `select`: `values`, `maxSelect`
- `file`: `mimeTypes`, `thumbs`, `maxSelect`, `maxSize`, `protected`

References:
- https://pocketbase.io/old/docs/go-collections
- https://pocketbase.io/old/docs/js-collections
- https://github.com/pocketbase/pocketbase/blob/v0.22.22/models/schema/schema_field.go

## Migration Notes (v0.23)
Key changes introduced in v0.23 migration:

- `schema` column renamed to `fields` in `_collections`
- Options flattened (legacy `options.*` mapped into top-level keys)
- System fields explicitly added to the field list

Reference:
- https://github.com/pocketbase/pocketbase/blob/master/migrations/1717233556_v0.23_migrate.go

## Code Generator Considerations

If your generator supports both formats:

1. Accept both `schema` and `fields` arrays
2. Normalize into a unified `fields` list
3. Handle `minSelect`/`maxSelect` even if they appear at the field top-level or inside `options`
4. Expect system fields (`id`, `created`, `updated`) to exist as explicit fields in v0.23+
5. For relation/select/file, map `maxSelect` > 1 to slices (`[]string`) and `maxSelect == 1` to single values

Local generator references:
- `internal/generator/schema.go`
- `internal/generator/mapper.go`
- `internal/generator/parser_test.go`
