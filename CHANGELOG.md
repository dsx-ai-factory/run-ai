# Changelog

Release notes for this project are published on the GitHub Releases page of this repository. Each tagged release has its notes generated automatically from the merged pull requests on that release branch.

For the list of currently supported versions, see [Support](SUPPORT.md).

## Versioning Policy

Releases are tagged `Ma.Mi.Pa`, following semantic versioning:

- **`Ma.Mi`** (major.minor, e.g. `3.0`) - A release branch (see [Release Process](RELEASE.md)). May carry new features, and is where breaking changes are introduced when unavoidable. Breaking changes are called out explicitly in that release's notes on the [GitHub Releases page](https://github.com/dsx-ai-factory/run-ai/releases).
- **`Pa`** (patch, e.g. `3.0.3`) - Fixes only, no new features and no breaking changes, cut from the corresponding `Ma.Mi` release branch.

For example: `3.0.0` is the first release on the `3.0` release branch and may include new features or breaking changes (documented in its release notes); `3.0.1` is a patch cut from that same `3.0` branch containing fixes only, with no new features and no breaking changes relative to `3.0.0`.

See [Support, Lifecycle, and Deprecation Policy](docs/about-enterprise/product-support-policy.md) for how long each `Ma.Mi` version receives patches.
