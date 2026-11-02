# Release Process

This document describes how NVIDIA Run:ai releases are cut, versioned, and
published from this repository.

## Cadence

Releases follow a minor-version cadence aligned with the versions tracked in [Support](SUPPORT.md). The latest 5 minor releases are supported at any given time; when a new minor release ships, the oldest supported branch moves to end-of-life.

## Branch Naming

- **`main`** - Active development. All feature work and fixes land here first.
- **Release branches** - Cut from `main` using the pattern `v<major>.<minor>` (e.g. `v3.0`) or `release-<major>.<minor>`. Creating a branch matching either pattern automatically triggers the release-branch automation (see [`.github/workflows/create-release-branch-oss.yml`](.github/workflows/create-release-branch-oss.yml)).
- Fixes that need to reach a supported release are backported via the `backport/v2.X` label (see [`.github/workflows/backport.yaml`](.github/workflows/backport.yaml) and [Maintainers](MAINTAINERS.md)).

## Tags

- Release tags use the pattern `X.Y.Z`; see [`.github/workflows/tag-trigger.yaml`](.github/workflows/tag-trigger.yaml).
- Tags **must** be created via a [GitHub Release](https://github.com/dsx-ai-factory/run-ai/releases), not pushed directly: the tag lane looks up the release's `target_commitish` and fails the run if no matching GitHub Release exists.
- Tags must target a release branch, not `main`/`master`. Releasing directly from `main`/`master` is rejected by CI.

## Who Can Trigger a Release

Making release decisions for supported versions is a maintainer responsibility (see [Maintainers](MAINTAINERS.md#responsibilities)). Maintainers create the release branch and the GitHub Release/tag; the automation described above then takes over publishing.

## What Happens on Release

Pushes to `main`, `master`, and release branches, and pushes of release tags, dispatch to the release and tag lanes ([`.github/workflows/release-trigger.yaml`](.github/workflows/release-trigger.yaml), [`.github/workflows/tag-trigger.yaml`](.github/workflows/tag-trigger.yaml)), which build and publish artifacts. Release notes are generated automatically from merged pull requests on the release branch and published to the [GitHub Releases page](https://github.com/dsx-ai-factory/run-ai/releases); see [Changelog](CHANGELOG.md).
