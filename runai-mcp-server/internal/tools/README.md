# Authoring Run:ai MCP Tools

This package is where every Run:ai MCP tool lives. The MCP server registers
nothing on its own — `tools.Register` is the single entry point, and every
new tool must be reachable from there.

This guide covers two things:

1. **General guidelines** — design principles for tools an LLM will call.
2. **Framework usage** — the `runaimcp` types, middleware, and conventions
   you must follow.

---

## 1. General guidelines

### Audience: an LLM, not a human

Tool names, descriptions, input fields, and output shapes are read by a model
deciding what to call and how to call it. Optimize for clarity, not brevity.

- **Name** — `snake_case`, verb-first when an action, noun-only when a fetch:
  `list_projects`, `get_workload`, `update_role`, `delete_access_rule`.
- **Description** — say *what it does*, *when to call it*, *when not to*
  (point at the sibling tool that fits instead), and *what it returns*, in
  one or two sentences.
- **Input fields** — name each field the way the upstream API spells it on
  the wire (`workloadId`, not `workload_id`), and accept enum values
  exactly as the API does, so the model can chain tools without
  translation. The handler forwards inputs as-is — no silent renames or
  value rewriting. Every field needs a `jsonschema:"..."` describing what
  it is; the lint test fails without one.
- **Outputs** — forward a compact subset of the upstream response under the
  same field names and types. No renames, conversions, or derived fields.

### Scope: one tool, one resource, one verb

Resist the urge to build "do everything for X" tools. The model can chain
calls; a wide tool with twenty optional inputs is harder for it to reason
about than five focused tools. One tool per resource × operation is the
default.

### Read-only vs. mutating

- Default to **read-only**. List/get/describe operations should set
  `ReadOnly: true`.
- Mutations (`create_*`, `update_*`, `delete_*`, `stop_*`) set `ReadOnly:
  false`. Destructive ones (delete, stop, revoke) additionally set
  `Destructive: true`.
- The server hides non-read-only tools unless started with
  `--allow-write-tools`. That gate is the user's safety net; don't fight it
  by mislabeling a mutating tool as read-only.

### Authorizing mutations

Write tools fail fast on a forbidden caller so a denied request never reaches the
upstream mutation endpoint (the upstream API stays the authoritative RBAC boundary;
this probe is a courtesy that turns a late 403 into an early one).

The pattern, shared via `common`:

1. `common.PermittedScopeForAction(ctx, client, resourceType, action)` fetches the
   caller's permitted scopes for the resource and action.
2. A scope matcher decides whether that scope covers the target —
   `common.ScopeAllowsCluster` for cluster-scoped resources (node pools),
   `common.ScopeAllowsOrgUnit` for projects/departments. Add a matcher to `common`
   rather than hand-rolling a third in a tool.
3. On no match, return `ErrForbidden`.

**Authorize as early as you know the target's scope identity.** When the input
carries it (e.g. a required `clusterId`), probe before doing anything else. When it
doesn't — an org-unit targeted by id, whose `clusterId` is only known after the GET —
probe right after that fetch. Don't force a generic "authorize before the handler"
step that can't see the resolved identity.

### Confirming mutations

A non-dry-run `create` / `update` / `delete` must be approved by a human before
it touches upstream. Route every mutation through `common.ConfirmWrite`, called
inside the handler *after* you've resolved the target and built its
before-snapshot / diff, and *before* the upstream write:

```go
if err := common.ConfirmWrite(ctx, in.Confirm, "delete", "node pool", name, id, out.Before); err != nil {
    return out, err
}
```

`ConfirmWrite` prefers a real human-in-the-loop prompt and degrades safely:

- **Client supports elicitation** — the server sends an `elicitation/create`
  prompt and proceeds only on *accept*. *Decline* / *cancel*, a disconnect, or a
  timeout aborts with `confirmation_required`. The model-supplied `confirm` flag
  is **ignored** on this path, so a model can't approve on the user's behalf.
- **No client can prompt** (headless / cron, or a client that didn't advertise
  form elicitation) — falls back to the `confirm=true` gate: `confirm=false`
  returns `confirmation_required` carrying the snapshot; `confirm=true` proceeds.

That contract means every mutating input needs a `Confirm bool` field (JSON
`confirm`), and the handler must honor `dryRun` — preview the before/after/diff,
with **no** prompt and no mutation — before it reaches `ConfirmWrite`.

The `details` map is rendered in the prompt (and the fallback error), so pass the
before-snapshot for a delete or the diff for an update. The prompt is
length-bounded and non-scrolling in some clients, so it's a summary — the full
before/after stays available via `dryRun`. To surface an extra warning at confirm
time (e.g. "deleting a project also deletes its workloads"), wrap the details with
`common.WithConfirmWarning`.

Elicitation is a *client* capability; the server advertises nothing for it and
only needs the code to call `Elicit`. Handlers reach the live session through
`ctx` (stashed by the framework); tests substitute a fake via
`runaimcp.WithSession`.

### Validation: fail loud, fail early

Validate everything you can before hitting the upstream API:

- Required fields are present and non-empty.
- Enum-like strings are in the allowed set.
- Pagination bounds are sane.
- Mutually-exclusive options aren't both set.

A clear `validation: projectId is required` is more useful than a generic
upstream `400 Bad Request`.

### Errors: stable categories beat verbose strings

The framework rewrites every error with a category prefix (`auth:`,
`forbidden:`, `not_found:`, `upstream:`, ...) so the LLM can pattern-match
without reading sentence-level English. To make that work, wrap upstream
errors with the matching sentinel — see [Error handling](#error-handling).

### Output shape: serializable, never nil

JSON `null` for an array is a frequent footgun. If you advertise
`pods: []nodeWorkloadPod` in the schema, the wire response must be `[]`
even when empty. Initialize slices explicitly (`Pods: []podEntry{}`) rather
than relying on `append` to a nil slice.

### Don't leak secrets

Tokens, client secrets, and signed URLs must not appear in tool outputs or
error messages. The recover middleware already scrubs panic details for this
reason; do the same in normal flow.

### Test what you write

At minimum, register the tool, run the lint test, and exercise validation
paths. For tools that touch upstream APIs, table-test the
status-code → sentinel mapping and the happy path output shaping.

---

## 2. Framework usage

The framework lives in `internal/runaimcp`. The contract:

- Tools are declared via `runaimcp.Tool[In, Out]`.
- They are registered via `runaimcp.Register`.
- Inputs must satisfy `runaimcp.ToolInput` (a `Validate() error` method).
- Upstream errors should wrap one of the `runaimcp.Err*` sentinels.

### File layout

Tools are grouped into **category subpackages**. Each category owns a
directory; each tool inside a category owns a file:

```text
internal/tools/
  tools.go               // central Register — wires every category in
  register_test.go       // lint test
  README.md              // this file
  common/                // cross-category helpers (UUID validation, etc.)
    validate.go
  identity/
    register.go          // identity.Register — one line per tool in the category
    whoami.go            // self-contained: types, Validate, registerWhoami
  workloads/
    register.go
    list_workloads.go
    get_workload.go
```

The contract per category package:

- `register.go` exports `func Register(r *runaimcp.Registry, c *runaiclient.Client)`
  that calls every `register<Name>` in the package.
- Each tool file declares its own input/output structs, `Validate`, and a
  `register<Name>` function that contains the full `runaimcp.Register(...)`
  call — types, metadata, and handler stay together as a single cohesive
  unit.

Adding a tool to an existing category = new `<name>.go` file + one line in
that category's `register.go`. Adding a new category = new subdirectory +
one line in the top-level `tools.go`.

Keep helpers private to the category package unless another category
genuinely shares them. Shared helpers go in `internal/tools/common/`
(exported) — `common.ValidateUUID` is the current example.

### Declaring a tool

`internal/tools/projects/get_project.go`:

```go
package projects

type getProjectInput struct {
    ProjectID string `json:"projectId" jsonschema:"Run:ai project ID"`
}

func (in getProjectInput) Validate() error {
    if strings.TrimSpace(in.ProjectID) == "" {
        return fmt.Errorf("projectId is required")
    }
    return nil
}

type getProjectOutput struct {
    ID   string `json:"id"   jsonschema:"Project ID"`
    Name string `json:"name" jsonschema:"Project name"`
}

func registerGetProject(r *runaimcp.Registry, client *runaiclient.Client) {
    runaimcp.Register(r, runaimcp.Tool[getProjectInput, getProjectOutput]{
        Name:        "get_project",
        Description: "Fetch a Run:ai project by ID.",
        ReadOnly:    true,
        Handler: func(ctx context.Context, in getProjectInput) (getProjectOutput, error) {
            resp, httpResp, err := client.BFF.ProjectsAPI.GetProject(ctx, in.ProjectID).Execute()
            if err != nil {
                return getProjectOutput{}, fmt.Errorf("get project %q: %w", in.ProjectID, runaimcp.MapHTTPError(httpResp, err))
            }
            return getProjectOutput{ID: resp.Id, Name: resp.Name}, nil
        },
    })
}
```

`internal/tools/projects/register.go`:

```go
package projects

func Register(r *runaimcp.Registry, client *runaiclient.Client) {
    registerGetProject(r, client)
    // registerListProjects(r, client)
}
```

Wire the category from the top-level `internal/tools/tools.go`:

```go
func Register(r *runaimcp.Registry, client *runaiclient.Client) {
    identity.Register(r, client)
    projects.Register(r, client)
    // ...
}
```

### Empty input

For list-style tools that take no parameters, declare an empty input struct
that satisfies `ToolInput`:

```go
type whoamiInput struct{}
func (whoamiInput) Validate() error { return nil }
```

### Tool flags

| Field           | When to set                                                       |
| --------------- | ----------------------------------------------------------------- |
| `ReadOnly`      | `true` for list/get/describe; `false` for any mutation            |
| `Destructive`   | `true` for delete/stop/revoke or any irreversible state change    |
| `MinAPIVersion` | tool relies on an endpoint introduced in a specific CP version    |
| `MaxAPIVersion` | tool relies on an endpoint removed in a later CP version          |

Version gating is enforced at registration time against the runtime
control-plane `/_version`. A tool outside the supported range is silently
skipped, so the LLM only sees tools that will actually work.

### Middleware (automatic, in order)

Every tool runs through:

1. **logging** — start/end log lines with the post-classify category.
2. **classify** — wrap the returned error with its category prefix (`auth:`,
   `forbidden:`, `not_found:`, `rate_limited:`, `upstream:`, `timeout:`,
   `canceled:`, `validation:`, `internal:`).
3. **recover** — panics become `ErrInternal`; stack traces go to the logger,
   never to the LLM.
4. **validate** — `in.Validate()`; a non-nil return short-circuits before
   the handler runs.
5. **handler** — your code.

Don't replicate these concerns inside your handler.

### Error handling

Wrap upstream failures with the right sentinel so the classifier can attach
the right LLM-facing category. The framework provides a helper for
generated-client errors:

```go
resp, httpResp, err := client.BFF.SomeAPI.DoThing(ctx, id).Execute()
if err != nil {
    return out, fmt.Errorf("do thing %q: %w", id, runaimcp.MapHTTPError(httpResp, err))
}
```

`MapHTTPError` maps:

| Status              | Sentinel               | Category prefix |
| ------------------- | ---------------------- | --------------- |
| 401                 | `ErrAuth`              | `auth`          |
| 403                 | `ErrForbidden`         | `forbidden`     |
| 404                 | `ErrNotFound`          | `not_found`     |
| 429                 | `ErrUpstreamThrottled` | `rate_limited`  |
| other non-2xx / nil | `ErrUpstream`          | `upstream`      |

If you need a sentinel without an HTTP context (e.g. a precondition check),
wrap directly:

```go
return out, fmt.Errorf("project %q has no nodes: %w", id, runaimcp.ErrNotFound)
```

`ErrValidation` is reserved for the validate middleware — return plain
errors from `Validate()` and they will be wrapped automatically.

### Wiring a generated client

Most Run:ai APIs already have a generated client:
- `runai-clients/<service>/generated/` — local, `require` + `replace`.
- `control-plane/<service>/pkg/client/generated/` — versioned, `go get`.

1. Add the generated package as a `go get` dependency.
2. Add a typed field on `Client` in `internal/runaiclient/client.go`
   (e.g. `Authorization *authzsvc.APIClient`).
3. Construct it in `runaiclient.New` using the **shared authenticated HTTP
   client** and the trimmed base URL — every generated client reuses the
   same `*http.Client` so bearer-token injection and 401 refresh come for
   free.

Call the generated client directly from the tool handler — don't use adapter wrapper inside `runaiclient`

Do not call generated-client methods at registration time. The lint test
runs `Register(r, nil)` to introspect schemas without invoking handlers; if
your `register*` function dereferences `client`, the test crashes.

### The lint test

`register_test.go` runs `runaimcp.Lint` against `Register(r, nil)`. It boots
an in-memory MCP server, calls `tools/list`, and verifies what the LLM would
actually see:

- Tool names match `^[a-z][a-z0-9]*(_[a-z0-9]+)*$`.
- Every input property has a non-empty `description` (from the
  `jsonschema:"..."` tag).
- Every output property has a non-empty `description` (same source).

Run it with `make test`. Adding a tool with a missing `jsonschema` tag is the
most common breakage; the test catches it before review.

### Testing a tool

Every tool ships with a `<name>_test.go` alongside its `<name>.go` file. The
lint test catches schema mistakes but never invokes a handler — the per-tool
test is what verifies the handler actually works.

**Required cases per tool:**

- **Validation** — one case per branch in `Validate()`, plus a happy case.
- **Handler happy path** — the output shape an LLM will see. Assert against
  the typed struct; check that slices/maps are non-nil where the schema
  promises them.
- **Error mapping** — one case per error category the handler can produce.
  Assert with `errors.Is(err, runaimcp.ErrAuth)` (or `ErrNotFound`, etc.),
  never substring-match the error message — the classify middleware will
  prefix it.

**Two helpers** wire the dependencies:

- `internal/auth/authtest.FakeProvider` — implements `auth.Provider` with a
  configurable token / token error. Use this directly when the handler only
  needs `client.Token(ctx)` (no BFF calls).
- `internal/runaiclient/runaiclienttest.NewServer(t)` — returns a
  `(*runaiclient.Client, *http.ServeMux)` pointed at an `httptest.Server`.
  Register canned upstream responses on the mux with `mux.HandleFunc(...)`.

**Pattern for a BFF-touching tool:**

```go
func TestListProjects(t *testing.T) {
    client, mux := runaiclienttest.NewServer(t)
    mux.HandleFunc("/api/v1/projects", func(w http.ResponseWriter, r *http.Request) {
        _ = json.NewEncoder(w).Encode(...) // fixture
    })
    got, err := listProjectsRun(context.Background(), client, listProjectsInput{...})
    // assertions...
}
```

Pull the handler body into an unexported `<tool>Run(ctx, client, in)` function
so the test can call it directly without going through the middleware stack —
middleware is covered by `runaimcp` tests. `whoami.go` / `whoami_test.go` is
the canonical example.

### Read-only mode

When the CLI runs without `--allow-write-tools`, the registry is constructed
with `ReadOnly: true` and silently drops every non-read-only tool. Two
implications:

- Default-on behavior should be safe to expose to anyone with a valid
  bearer token. If your tool needs the write flag to be useful, mark it
  `ReadOnly: false` honestly.
- If a mutating tool is missing from `tools/list`, check the flag before
  digging into registration code.

---

## Quick checklist

Before opening a PR for a new tool:

- [ ] Snake-case tool name; description says what it does, when to call it,
      and when not to (point at the sibling tool that fits instead).
- [ ] Input/output field names match the upstream API's JSON spelling;
      input enum values match what the API accepts on the wire.
- [ ] `Validate() error` rejects bad input with a useful message.
- [ ] Output struct is compact and JSON-friendly; slices initialized non-nil.
- [ ] `ReadOnly` / `Destructive` set correctly.
- [ ] Mutating tools: a `Confirm bool` input, `dryRun` honored, and the
      upstream write gated behind `common.ConfirmWrite`.
- [ ] `MinAPIVersion` / `MaxAPIVersion` set if the endpoint is version-bound.
- [ ] Upstream errors wrapped with `runaimcp.MapHTTPError` (or a direct
      sentinel where appropriate).
- [ ] Tool wired from its category's `register.go`, and the category wired
      from the top-level `tools.Register`.
- [ ] `<name>_test.go` covers validation branches, the handler happy path,
      and every error category the handler can produce.
- [ ] `make test` passes.
