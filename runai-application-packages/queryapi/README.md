# queryapi

`queryapi` is a Go library in the Run:ai `runai-application-packages` tree that
turns the platform's URL query parameters (`filterBy`, `sortBy`) into safe,
validated database query fragments. It gives list endpoints a single, consistent
way to parse client-supplied filters/sorts, map API field names to real DB
columns, validate values, and translate operators into SQL. It is a **library
package**: no binary, no Dockerfile, consumed by other modules through `go.mod`.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Public API](#public-api)
- [Importing](#importing)
- [Building and testing](#building-and-testing)

## Overview

The root package (`package queryapi`) parses the raw query strings and the
`dbenrich` subpackage applies the parsed result to an ORM query:

- **Parsing (`parser.go`)** — `Parser`, built from an `APIFieldsMapping`, splits
  each `filterBy` string of the form `<field><operator><value>` using the
  supported operators (`==`, `!=`, `<=`, `>=`, `=@` contains, `!@` not-contains,
  `=^` starts-with, `=$` ends-with). Only fields present in the mapping are
  accepted; each maps to a DB `Column`. Multiple values are OR-joined with `||`
  (capped at 10 options). Operators are translated to SQL (e.g. `=@` → `ILIKE`
  with `%value%`, `!=` → `is distinct from`).
- **Validation (`validator.go`)** — every field mapping carries `Validator`s;
  `UserInputValidator` (length cap) is always added, and `UUIDValidator` /
  `DateValidator` can be attached per field via `NewFieldMapping`.
- **Field/sort mapping (`types.go`)** — `GetMapping` resolves an API `sortBy`
  field to its DB column.
- **DB enrichment (`dbenrich/`)** — `Filters([]Filter)` and `SortBy(column,
  order)` return `orm.Enricher` functions that apply `WHERE`/`ORDER BY` clauses
  onto a `runaidb/orm` query, with `NULLS FIRST/LAST` ordering semantics.

## Repository layout

```text
.
├── parser.go              # Parser, operator table, filter parsing
├── parser_test.go
├── types.go               # APIFieldsMapping, FieldMapping, Filter, Sort
├── validator.go           # Validator interface, UUID/Date/UserInput validators
├── dbenrich/              # ORM enrichers (Filters, SortBy) over runaidb/orm
│   ├── filters.go
│   ├── sort.go
│   └── *_test.go
├── integration-tests/     # Separate Go module; DB-backed filter/sort tests + migrations
├── artifact-metadata.json
├── Makefile
└── go.mod
```

## Public API

| Symbol | Description |
|--------|-------------|
| `NewParser(APIFieldsMapping) *Parser` | Build a parser from field→column mappings |
| `NewFieldMapping(column, ...Validator) FieldMapping` | Declare a mapped field (always adds `UserInputValidator`) |
| `(*Parser).Filters([]string) ([]Filter, error)` | Parse `filterBy` params into `Filter`s |
| `(*Parser).GetMapping(field) (string, error)` | Resolve a `sortBy` field to a DB column |
| `Validator` + `UUIDValidator` / `DateValidator` / `UserInputValidator` | Per-field value validation |
| `dbenrich.Filters([]Filter) orm.Enricher` | Apply parsed filters as `WHERE` clauses |
| `dbenrich.SortBy(column, order) orm.Enricher` | Apply an `ORDER BY` clause |

## Importing

```go
import (
    "github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
    "github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi/dbenrich"
)

mapping := queryapi.APIFieldsMapping{
    "name":      queryapi.NewFieldMapping("workload_name"),
    "createdAt": queryapi.NewFieldMapping("created_at", queryapi.DateValidator{}),
}
parser := queryapi.NewParser(mapping)
filters, err := parser.Filters([]string{"name=@train"})
db = dbenrich.Filters(filters)(db)
```

## Building and testing

The module only `include`s the shared makefiles under
[`build/makefile`](../../build/makefile); it is a library, so there is no image
to build (`make build` is a no-op).

| Target | Description |
|--------|-------------|
| `make test` | Run unit tests (`test-go`) **and** the integration tests |
| `make integration-tests` | `cd integration-tests && go test -v ./...` — DB-backed tests |
| `make lint` | Run `golangci-lint` (`lint-go`) in a linter container |
| `make validate` | Package import/replace checks (no build step) |
| `make build` | No-op — this is a library, not a deployable artifact |

`integration-tests/` is its own Go module and exercises the enrichers against a
real database (migrations under `integration-tests/migrations/up`). Unit tests
run inside a Docker `golang:<version>` image (version from [`go.mod`](go.mod)).
