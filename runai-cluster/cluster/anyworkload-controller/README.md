# anyworkload-controller

This module is a thin **consumer** of the [`dromos`](../dromos) library. It is
the binary Run:ai ships in clusters: it imports dromos, imports the
runai-specific `networking` and `personalcreds` plugins (sub-packages of this
module) and the generic `mnnvl` and `storage` plugins (sub-packages of
dromos), and starts the controller via `app.RunController(ctx, builders)`.

```
.
├── cmd/main.go         calls app.RunController with the chosen builders
├── go.mod              requires dromos via local replace
├── plugins/
│   ├── networking/     runai-specific plugin
│   └── personalcreds/  runai-specific plugin
├── tests/              envtest integration suite
├── Dockerfile          ubi-minimal, multi-arch
└── Makefile            build + docker-build + envtest
```

## Why a thin consumer

The `dromos` library is the shipping product. Run:ai's binary is one consumer
of it. A third party writes the same shape of `main.go` in their own repo -
imports dromos, adds their own plugin builders, calls
`app.RunController(ctx, builders)`. No fork of this repo required.

In-tree plugins (`networking`, `personalcreds` here, `mnnvl`, `storage` in
dromos) are sub-packages of their host module - they ship with their host's
release cycle. Third-party plugins live in their own repo with their own
`go.mod` and import dromos directly.

When dromos is open-sourced the `replace ../dromos` in this `go.mod` drops
away and dromos resolves through its public module path.

## Adding a plugin

To add a runai-specific plugin to this binary:

1. Drop a new sub-package under `plugins/`. Copy the shape of `plugins/networking/`.
2. Append the plugin's `New` to the `builders` slice in `cmd/main.go`.

The full plugin-authoring guide lives in
[`../dromos/PLUGIN_AUTHORING.md`](../dromos/PLUGIN_AUTHORING.md). The in-tree
plugins ([`plugins/networking/`](plugins/networking),
[`plugins/personalcreds/`](plugins/personalcreds),
[`../dromos/plugins/mnnvl/`](../dromos/plugins/mnnvl),
[`../dromos/plugins/storage/`](../dromos/plugins/storage)) are the canonical
references.

## Build & test

```bash
make build           # produces bin/anyworkload-controller-<arch>
make unittest        # unit tests for this module + plugins/
make test-dev        # full envtest integration suite
make docker-build    # builds the container image
```

The dromos library has its own unit tests under `../dromos`. Run them from
there.
