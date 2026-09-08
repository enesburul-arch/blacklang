# File and Image Media Fields

BlackLang v0.2 supports local `file` and `image` fields as stored scalar values for generated web apps.

```black
entity Product {
  photo image optional accept "image/*" label "Product Photo"
  specSheet file optional accept "application/pdf" label "Spec Sheet"
}

page Products {
  source Product

  table {
    columns photo, specSheet
  }

  form {
    fields photo, specSheet
  }
}
```

Generated React forms render file inputs. When a user selects a file, the generated page reads it as a data URL and submits that string through the normal JSON API path.

Generated output:

- stores `file` and `image` fields as strings in Prisma and SQLite/PostgreSQL schemas
- renders `image` values as thumbnails in table/detail/form preview UI
- renders `file` values as links
- validates image values as `data:image/*` or `http(s)` URLs
- validates file values as `data:*` or `http(s)` URLs
- emits OpenAPI `x-blacklang-media` and `x-blacklang-encoding: data-url` metadata
- increases generated JSON body size to `2mb` when an entity has media fields

`accept` is optional and valid only on `file` or `image` fields. `image` defaults to `accept "image/*"` when no accept value is declared. For non-image documents, use `file accept "application/pdf"` or another MIME hint.

Do not put storage provider secrets, bucket names, upload tokens, or private endpoints directly in `.black` files. External storage adapters are future provider behavior; this MVP is deterministic local data URL storage.

Custom action media inputs and auth user media fields are not generated in this MVP. Model media as entity fields and expose them through generated page forms or explicit API body contracts.

Useful commands:

```bash
black docs media --json
black explain media --json
black inspect examples/warehouse/app.black --affected Product.photo --json
black validate examples/warehouse/app.black --json
black build examples/warehouse/app.black --out generated --json
```
