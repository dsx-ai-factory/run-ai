# runai-application-packages

This repository contains versioned Run.ai packages. Each package included here is under the respirability of the team that created it.
Desired permission to your package should be managed using the `CODEOWNERS` file.

## Automated Release Process

This repository provides automated Make commands to simplify the versioning process. Each package directory supports the following commands:

### Production Release

```bash
cd <package-name>
make check-breaking-changes      # optional: preview API diff vs last prod tag
make package-bump-version        # automatic: patch (compatible) or minor (breaking)
```

**Requirements:**
- Must be on `master` branch
- Local `master` must be up-to-date with `origin/master`

**Behavior:**
- Compares **only this package** to its latest `${TAG_FORMAT}/vX.Y.Z` prod tag
  (e.g. `runai-application-packages/async-events/v1.0.6`) using Go `apidiff`
- `make package-bump-version` automatically chooses **patch** when there are no
  breaking public API changes (including backward-compatible additions), or
  **minor** when breaking public API changes are found
- There is no makefile major bump target
- Increments the version from the latest production and development tags and pushes
  the new tag

**Example:**
```bash
cd async-events
make package-bump-version
# Output: Published tag: async-events/v1.0.7
```

### Development Release

```bash
cd <package-name>
make package-bump-version-dev
```

**Requirements:**
- Can be run from any branch

**Behavior:**
- Checks **both** the latest prod tag and the latest dev tag, and computes a
  prod-derived candidate version: next patch/minor off the latest prod tag
  (patch vs minor decided via `apidiff`, same as `package-bump-version`)
- **If the latest dev tag's version is already ahead of (or equal to) that
  candidate** (e.g. dev is `v1.3.0-dev.2` while prod is only `v1.2.0`):
  continues the dev sequence at the dev tag's own version, e.g. `v1.3.0-dev.3`
- **Otherwise** (prod is more advanced than the existing dev tag, or there is
  no dev tag yet): starts a fresh `-dev.0` at the candidate version, e.g.
  `v1.2.4-dev.0` or `v1.3.0-dev.0`
- **If no tags exist at all**: Creates `v0.0.0-dev.0`

**Example:**
```bash
cd async-events
make package-bump-version-dev
# Output: Published tag: async-events/v1.0.6-dev.1
```
<br/><br/>

## Manual Release Process

To release a Go package, follow these steps:
> [!IMPORTANT]  
> This is a production release, for a dev release check out [here](#development-release-process)

1. **Tag the Branch**:  
   To release a package, you need to tag the master branch using the following format:

   ```
   package_folder_name/vX.X.X
   ```

   - Replace `package_folder_name` with the actual directory name where the package resides.
   - `vX.X.X` should be replaced with the desired version number (following semantic versioning rules).

   Example tag:

   ```
   logging/v1.2.3
   ```

2. **Import the Package**:  
   Once the branch is tagged, it can be imported by other services or repositories using the standard Go module import syntax.

3. **Follow Versioning Rules**:  
   When tagging a release, it's important to adhere to [semantic versioning](https://semver.org/) guidelines to maintain consistency and stability across your packages. The format for versioning is `vMAJOR.MINOR.PATCH`.

   - **MAJOR**: Increases when incompatible API changes are introduced.
   - **MINOR**: Increases when functionality is added in a backward-compatible manner.
   - **PATCH**: Increases when backward-compatible bug fixes are introduced.

   Example:
    ```yaml
   - v1.0.0
   - v2.1.0
   - v2.1.1
    ```
<br/><br/>

## Development Release Process
**A dev tag can be preformed from any branch**

To release a development version of a package, use the same tagging process as above but append `-dev.x` to the version number.

For example, to release a development version, tag the branch as follows:

```
package_folder_name/vX.X.X-dev.0
```

Example:
```
logging/v1.2.3-dev.0
```

This tag will mark the package as a development release, indicating that it is a work-in-progress version, not yet intended for production use.

### Versioning Development Releases

Development releases should also follow proper versioning guidelines. However, the `-dev.x` suffix signals that the release is not considered stable and may still undergo changes before the official release.

- Use `vX.X.X-dev.x` for development versions leading up to a full release.
- Increment `DEV` version, until you ready to release the next patch/minor (e.g., `v1.2.3-dev.1`).
- Once the package is stable, remove the `-dev.x` suffix and create the final release tag (e.g., `v1.2.3`).
<br/><br/>
