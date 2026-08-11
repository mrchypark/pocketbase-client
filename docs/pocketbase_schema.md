# PocketBase Schema Format

This document captures the **schema JSON structure and field options** needed by the code generator. The generator supports the latest PocketBase schema format (v0.23+): collections expose their fields under the `fields` key and field options are flattened onto the field itself (no nested `options` object).

## Collection Object (Top-Level Fields)

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

## Field Objects — Common Keys

All fields share the following properties:

- `id` (string)
- `name` (string)
- `type` (string)
- `system` (boolean)
- `hidden` (boolean)
- `presentable` (boolean)
- `required` (boolean)

Reference:
- https://github.com/pocketbase/pocketbase/blob/master/core/fields_list.go

## Field Objects — Type-Specific Keys

The following keys are **top-level** (flattened; no `options` object). These come directly from the field structs in `core/field_*.go`.

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
| `geoPoint` | https://github.com/pocketbase/pocketbase/blob/master/core/field_geo_point.go | `required` (added in v0.39.0) |
| `email` | https://github.com/pocketbase/pocketbase/blob/master/core/field_email.go | `exceptDomains`, `onlyDomains`, `required` |
| `url` | https://github.com/pocketbase/pocketbase/blob/master/core/field_url.go | `exceptDomains`, `onlyDomains`, `required` |
| `editor` | https://github.com/pocketbase/pocketbase/blob/master/core/field_editor.go | `maxSize`, `convertURLs`, `required` |
| `password` | https://github.com/pocketbase/pocketbase/blob/master/core/field_password.go | `pattern`, `min`, `max`, `cost`, `required` |

All field types also expose a `help` string key (field description, added in v0.39.0).

## JSON Example

From the API (`fields` array with flattened keys):

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

## Code Generator Considerations

The generator only accepts the latest format:

1. Collections must expose fields under the `fields` array
2. Field options are read from the field top-level (flattened keys); a nested `options` object is ignored
3. `minSelect`/`maxSelect` are always read from the field top-level
4. System fields (`id`, `created`, `updated`) exist as explicit fields and are skipped during generation
5. For relation/select/file, `maxSelect` > 1 maps to slices (`[]string`) and `maxSelect == 1` to single values

The schema input can be provided in any of the shapes PocketBase produces:

- The raw paginated response of `GET /api/collections` (`{"items": [...], "page": 1, "perPage": 30, ...}`)
- A plain JSON array of collection objects
- A single collection object (the response of `GET /api/collections/{id}`)

Local generator references:
- `internal/generator/schema.go`
- `internal/generator/mapper.go`
- `internal/generator/parser.go`
- `internal/generator/parser_test.go`
