# CodeRabbit Security Configuration

## Overview

CodeRabbit is an AI-powered code review tool installed as a GitHub App on the `run-ai` organization. It automatically reviews pull requests and provides feedback on code quality, security, and best practices.

This repo uses a `.coderabbit.yaml` file at the root to customize CodeRabbit's behavior on top of the organization-level defaults.

## How Configuration Works

CodeRabbit resolves configuration in this priority order:

1. **Repository YAML** (`.coderabbit.yaml` in this repo) -- highest priority
2. **Organization UI** (configured at [app.coderabbit.ai](https://app.coderabbit.ai))
3. **CodeRabbit schema defaults**

Our config sets `inheritance: true`, which **deep-merges** repo settings on top of org defaults rather than replacing them. This means:

- Settings we define override the org value
- Settings we don't define are inherited from the org
- Arrays (like path filters) are merged: our items first, then org items

## What the Org Config Provides

The organization-level config (managed by org admins at [app.coderabbit.ai](https://app.coderabbit.ai)) provides:

- **Review profile**: `chill` (lighter, fewer nitpicks)
- **Auto-review**: Enabled on all non-draft PRs
- **50+ static analysis tools** enabled, including:
  - `golangci-lint` -- Go linting
  - `gosec` -- Go security scanner
  - `hadolint` -- Dockerfile linting
  - `shellcheck` -- Shell script analysis
  - `gitleaks` / `trufflehog` -- Secret detection
  - `semgrep` / `opengrep` -- Multi-language security scanning
  - `checkov` / `trivy` -- Infrastructure-as-code and container security
  - `actionlint` -- GitHub Actions validation
  - `yamllint` -- YAML validation
  - `eslint` / `biome` -- JavaScript/TypeScript
  - `markdownlint` -- Markdown
- **PR summaries**: Auto-generated in PR descriptions
- **Suggested labels and reviewers**: Enabled (suggestions only, not auto-applied)
- **Knowledge base**: Learning from repo patterns
- **Finishing touches**: Docstring and unit test generation available via PR comments

## What Our Repo Config Adds

### Tone

Reviews are tuned to be concise and security-focused. SQL injection patterns are flagged as critical. Stylistic nitpicks are skipped.

### Path Filters

The following directories are **excluded from review** (auto-generated or build output):

| Pattern | Why |
|---------|-----|
| `**/swagger/_build/**` | Built/bundled OpenAPI output (26 services) |
| `**/generated/**` | OpenAPI codegen output (50+ directories across control-plane, runai-clients, e2e) |
| `**/vendor/**` | Go vendor dependencies |
| `**/node_modules/**` | npm dependencies |
| `**/*.generated.go` / `**/*.gen.go` | Generated Go files |
| `**/redoc/_build/**` | ReDoc build output |

Only human-written code is reviewed. For swagger, `swagger/source/` is reviewed while `swagger/_build/` is not.

### Path Instructions

Custom review guidance is applied based on file path:

#### Go Database Code (`control-plane/**/dao/**/*.go` and `control-plane/**/database/**/*.go`)

SQL injection detection for GORM. CodeRabbit is instructed to flag:

- **UNSAFE**: `fmt.Sprintf("%s > %s", column, value)` -- value interpolated directly
- **UNSAFE**: `db.Raw(query)` / `db.Exec(query)` where query is built with `fmt.Sprintf` and values embedded without `?` placeholders
- **UNSAFE**: String concatenation (`+`) to build WHERE clauses with data values

And to accept:

- **SAFE**: `fmt.Sprintf("%s > ?", column), value` -- column from constant, value as bind parameter
- **SAFE**: `db.Raw(sql, params...)` -- values passed as separate arguments

This is based on the [SQL Injection Retro](https://runai.atlassian.net/wiki/spaces/EN/pages/4518969345/SQL+Injection+Retro) where a real vulnerability was found in the assets service.

#### Dockerfiles (`**/Dockerfile`)

- Ensure non-root `USER` is set in the final stage
- Base images must use pinned versions or digests, not `:latest`

#### OpenAPI Specs (`control-plane/**/swagger/source/**`)

- Check for proper authentication requirements on endpoints
- Validate input schemas (maxLength, pattern, enum)
- Flag sensitive data exposure in responses

#### Test Files (`**/*_test.go`)

- Enforce parameterized queries (`?` placeholders) even in test code
- Prevents copy-paste of unsafe SQL patterns from tests into production

## Interacting with CodeRabbit

Developers can interact with CodeRabbit by commenting on PRs:

| Command | What it does |
|---------|-------------|
| `@coderabbitai review` | Trigger a review (useful for draft PRs or re-reviews) |
| `@coderabbitai resolve` | Mark all review comments as resolved |
| `@coderabbitai configuration` | Show the effective configuration for this repo |
| `@coderabbitai help` | List all available commands |
| `@coderabbitai generate docstrings` | Generate missing docstrings |
| `@coderabbitai generate unit tests` | Generate unit tests for changed code |

## Validation

All configuration changes were validated with dedicated test PRs:

| Test | PR | What was validated |
|------|-----|-------------------|
| Path filters | [#18869](https://github.com/dsx-ai-factory/run-ai/pull/18869) | `swagger/_build/` filtered out, `swagger/source/` reviewed |
| SQL injection | [#18872](https://github.com/dsx-ai-factory/run-ai/pull/18872) | 3 unsafe GORM patterns flagged, 2 safe patterns not flagged |
| Dockerfile | [#18873](https://github.com/dsx-ai-factory/run-ai/pull/18873) | `:latest` and missing `USER` flagged |
| OpenAPI spec | [#18874](https://github.com/dsx-ai-factory/run-ai/pull/18874) | Missing auth, missing validation, exposed secrets flagged |
| Test files | [#18875](https://github.com/dsx-ai-factory/run-ai/pull/18875) | Unsafe SQL in `_test.go` files flagged |

## Future Improvements

- **Custom ast-grep rules**: Structural pattern matching for `db.Raw(fmt.Sprintf(...))` without bind args -- more precise than natural language instructions
- **Enable gosec in all services**: Currently commented out in `org-unit-service` and others; separate from CodeRabbit but complementary
- **Additional path_instructions**: Could add rules for Helm charts, GitHub Actions workflows, or frontend security patterns as needed
