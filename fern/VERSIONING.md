<!--
SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->

# Versioned Docs

One Fern build of `main` publishes every supported version:

| Version | Nav file in `fern/docs.yml` | Content comes from |
|---|---|---|
| Newest (built from `main`) | `products/<product>.yml` | `main` |
| Superseded, still supported | `versions/v<X.Y>/fern/products/<product>.yml` | tip of `release-<X.Y>` |
| End of life | `archived-versions.yml` + redirect | externally hosted docs |

`fern/versions/` is generated and gitignored.
`tools/materialize-fern-versions.sh` reads which versions `docs.yml` registers
and fills `fern/versions/v<X.Y>/` from each `release-<X.Y>` branch, mirroring
the repo layout so the branch's `../../docs/...` nav paths resolve unchanged.
Nothing is synced back to `main`: every build reads the branch tip, so a docs
fix merged to `release-<X.Y>` is live on the next publish of `main`.

Where it runs:

- `fern-docs-ci.yaml`, before the nav check and `fern check`. It also runs on
  docs pushes to `release-*.*`, validating `main` against the new branch tip.
- The `run-ai/saas-configuration` preview and publish lanes, before
  `fern generate`.
- Locally, before `fern check` or `fern docs dev`:

  ```bash
  tools/materialize-fern-versions.sh --fetch
  python3 ./tools/check-fern-nav.py fern
  ```

## Cutover: When `<X.Y+1>` Docs Go Public

Until the cutover, the newest entry is labeled `v<X.Y>` and built from `main`,
so `<X.Y+1>` docs merged to `main` publish under that label. Do the cutover in
one PR to `main` when the `<X.Y+1>` docs should go public; `release-<X.Y>` must
already exist.

1. In `fern/docs.yml`, for **each** product (`self-hosted`, `multi-tenant`):
   - Repoint the `v<X.Y>` entry from `products/<product>.yml` to
     `versions/v<X.Y>/fern/products/<product>.yml`.
   - Add a `v<X.Y+1>` entry at the top of `versions:` with
     `path: products/<product>.yml`. The first entry is the default version.

   ```yaml
   versions:
     - display-name: "v<X.Y+1>"
       slug: v<X.Y+1>
       path: products/self-hosted.yml
     - display-name: "v<X.Y>"
       slug: v<X.Y>
       path: versions/v<X.Y>/fern/products/self-hosted.yml
   ```

2. Verify locally:

   ```bash
   tools/materialize-fern-versions.sh --fetch
   python3 ./tools/check-fern-nav.py fern
   (cd fern && fern check && fern docs dev)
   ```

   In the preview, open a page under `/<product>/v<X.Y>/` and confirm its links
   stay under `/<product>/v<X.Y>/`.

3. Make sure `release-<X.Y>` carries `tools/materialize-fern-versions.sh`,
   `tools/check-fern-nav.py`, and the current `fern-docs-ci.yaml`. A `push`
   workflow runs from the pushed branch's copy, so without them, docs pushes
   to `release-<X.Y>` are not validated against `main`. Branches cut from `main`
   after these files landed already have them.

## End of Life

When `<X.Y>` reaches EOL, point its `docs.yml` entries at
`archived-versions.yml` and add redirects for `/<product>/v<X.Y>/:slug*`, as
the `v2.2x` entries do. The script stops materializing it automatically.

## Known Limitation

The API reference tab (`api-name: main-api` / `multi-tenant-api`) resolves to
`fern/apis/` on `main` for every version, so superseded versions show the
newest API spec.
