# Relation Load Policy

Relation load policy controls which generated API response contexts eagerly include a related record object for an entity reference field.

## Syntax

```black
entity Customer {
  name text required
}

entity Order {
  customer Customer required load detail query label "Customer"
}
```

Use `load` only on relation fields, where the field type is another declared entity.

Supported scopes:

| Scope | Generated response context |
| --- | --- |
| `list` | Base page collection route, such as `GET /api/orders` |
| `detail` | Item read route, such as `GET /api/orders/{id}` |
| `query` | Bound custom query list route, such as `GET /api/orders/query` |
| `mutation` | Create, update, archive, restore, custom action, and workflow responses |
| `none` | No generated response includes the relation object |

Omitting `load` keeps the backward-compatible default: relation objects are included in list, detail, query, and mutation responses.

Use `load none` by itself when generated responses should return only the relation ID field, such as `customerId`.

## Generated behavior

The generator maps relation load scopes to deterministic response attachment. List and query routes collect relation IDs and load each relation target with one batch `findMany` per relation field; detail and mutation responses use the same helper for a single-item array. The database schema does not change. Forms still submit relation IDs, and relation select options still load from the related entity list route.

When roles and field permissions exist, generated routes attach a relation object only when the current role can read the source relation field. Attached relation records are sanitized with the target entity's read field policy before the response is returned.

Generated table, search, filter, sort, and detail rendering use the related record label when the relation object is loaded. If it is not loaded, generated UI falls back to the relation ID field, so rendering stays deterministic.

Generated OpenAPI operations expose:

```json
{
  "x-blacklang-relation-load-context": "query",
  "x-blacklang-relation-load": [
    {
      "field": "customer",
      "target": "Customer",
      "contexts": ["detail", "query"],
      "explicit": true,
      "included": true
    }
  ]
}
```

## Limits

Relation load policy does not add joins to custom query predicates. `query where`, `sort`, and `aggregate` still use stored primitive fields only.

`load none` cannot be combined with other scopes. Repeat neither the `load` modifier nor an individual scope.

## Diagnostics

| Code | Repair |
| --- | --- |
| `UNSUPPORTED_RELATION_LOAD_FIELD` | Move `load` to an entity reference field. |
| `MISSING_RELATION_LOAD_SCOPE` | Add `list`, `detail`, `query`, `mutation`, or `none`. |
| `UNSUPPORTED_RELATION_LOAD_SCOPE` | Use one of the supported scopes. |
| `DUPLICATE_RELATION_LOAD_SCOPE` | Keep each scope once. |
| `DUPLICATE_RELATION_LOAD` | Keep one `load` modifier on the field. |
| `CONFLICTING_RELATION_LOAD_SCOPE` | Use `load none` by itself, or remove `none`. |

Use `black docs relation-load --json`, `black explain relation-load --json`, and `black inspect app.black --affected Entity.relationField --json` before editing relation load behavior.
