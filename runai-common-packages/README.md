# runai-common-packages

This repository contains versioned Run.ai packages. The packages included here are infrastructure-level components that are used across Run.ai.

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
  (e.g. `runai-common-packages/logging/v1.2.3`) using Go `apidiff`
- `make package-bump-version` automatically chooses **patch** when there are no
  breaking public API changes (including backward-compatible additions), or
  **minor** when breaking public API changes are found
- There is no makefile major bump target
- Increments the version from the latest production and development tags and pushes
  the new tag

**Example:**
```bash
cd logging
make package-bump-version
# Output: Published tag: logging/v1.2.4
```

### Development Release

```bash
cd <package-name>
make package-bump-version-dev
```

**Requirements:**
- Can be run from any branch

**Behavior:**
- **If the latest tag is a dev tag** (e.g., `v1.2.3-dev.5`): Automatically increments the dev version to `v1.2.3-dev.6` without prompting
- **If the latest tag is a production tag** (e.g., `v1.2.3`): Prompts for minor/patch bump, then creates a new dev tag (e.g., `v1.2.4-dev.0` or `v1.3.0-dev.0`)
- **If no tags exist**: Creates `v0.0.0-dev.0`

**Example:**
```bash
cd cache
make package-bump-version-dev
# Output: Published tag: cache/v1.2.3-dev.6  (no prompt if latest is already a dev tag)
```

### Workflow Example

**Developing a Feature:**

```bash
# 1. Create and work on a feature branch
git checkout -b new-feature

# 2. Tag dev versions as you iterate
cd metrics
make package-bump-version-dev  # metrics/v1.2.3-dev.0
# ... make changes ...
make package-bump-version-dev  # metrics/v1.2.3-dev.1 (auto-increments, no prompt)
make package-bump-version-dev  # metrics/v1.2.3-dev.2 (auto-increments, no prompt)

# 3. After PR is merged to master
git checkout master
git pull

# 4. Create production release
make check-breaking-changes    # optional preview
make package-bump-version      # patch if compatible, minor if breaking
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
