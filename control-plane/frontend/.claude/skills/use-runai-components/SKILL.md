---
name: use-runai-components
description: >
  Enforces use of RunAI wrapper components instead of raw Quasar components when
  writing or editing Vue files in control-plane/frontend. Apply this
  whenever implementing, editing, or reviewing frontend Vue components.
user-invocable: false
---

# Use RunAI Components Instead of Raw Quasar

When writing or editing Vue components in `control-plane/frontend`, **always prefer the RunAI wrapper component** over the raw Quasar equivalent. Only fall back to the Quasar component directly if no RunAI wrapper exists.

## Component Map

| Instead of (Quasar) | Use (RunAI) | Import path |
|---|---|---|
| `q-btn` | `runai-btn` | `@/components/common/runai-btn` |
| `q-input` | `runai-input` | `@/components/common/runai-input` |
| `q-select` | `runai-select` | `@/components/common/runai-select` |
| `q-select` (autocomplete) | `runai-autocomplete-select` | `@/components/common/runai-autocomplete-select` |
| `q-toggle` | `runai-toggle` | `@/components/common/runai-toggle` |
| `q-table` | `runai-table` | `@/components/common/runai-table` |
| `q-tooltip` (standalone) | `runai-tooltip-wrapper` | `@/components/common/runai-tooltip-wrapper` |
| `q-tooltip` (icon-only btn) | `runai-tooltip` | `@/components/common/runai-tooltip` |
| `q-dialog` | `runai-base-dialog` | `@/components/common/runai-base-dialog` |
| `q-expansion-item` | `runai-expansion-item` | `@/components/common/runai-expansion-item` |
| `q-stepper` | `runai-stepper` | `@/components/common/runai-stepper` |
| `q-separator` (with label) | `runai-seperator-with-text` | `@/components/common/runai-seperator-with-text` |
| `q-date` / date picker | `runai-datepicker` | `@/components/common/runai-datepicker` |
| `q-chip` | `runai-filter-chip` | `@/components/common/runai-filter-chip` |
| `q-editor` | `runai-editor` | `@/components/common/runai-editor` |
| Delete confirmation dialog | `runai-delete-modal` | `@/components/common/runai-delete-modal` |
| Rename dialog | `runai-rename-modal` | `@/components/common/runai-rename-modal` |

## RunAI-Only Components (no Quasar equivalent — use these for common patterns)

| Use case | RunAI component | Import path |
|---|---|---|
| Page-level table + filters | `runai-table-wrapper` | `@/components/common/runai-table-wrapper` |
| Form with unsaved-changes guard | `runai-form-wrapper` | `@/components/common/runai-form-wrapper` |
| Form action buttons (save/cancel) | `runai-form-footer` | `@/components/common/runai-form-footer` |
| Form section card | `runai-form-card-section` | `@/components/common/runai-form-card-section` |
| Section title + divider | `runai-section` | `@/components/common/runai-section` |
| Dropdown button with menu | `runai-dropdown` | `@/components/common/runai-dropdown` |
| Copy-to-clipboard button | `runai-copy-to-clipboard-btn` | `@/components/common/runai-copy-to-clipboard-btn` |
| Code display block | `runai-code-block` | `@/components/common/runai-code-block` |
| Status indicator | `runai-status` | `@/components/common/runai-status` |
| SVG icon | `runai-svg-icon` | `@/components/common/runai-svg-icon` |
| Side drawer / panel | `runai-side-drawer` | `@/components/common/runai-side-drawer` |
| Multi-select with add | `runai-addable-multi-select` | `@/components/common/runai-addable-multi-select` |
| Radio button group | `runai-radio-options` | `@/components/common/runai-radio-options` |
| Checkbox list with search | `runai-checkbox-list` | `@/components/common/runai-checkbox-list` |

## Rules

1. **Always check this map first.** If a RunAI wrapper exists, use it — do not use the raw Quasar component.
2. **Imports**: import from the `@/components/common/<runai-component>` barrel (index file), not the `.vue` file directly.
3. **If no wrapper exists** for the Quasar component you need, use Quasar directly and leave a comment `<!-- no RunAI wrapper -->`.
4. **When in doubt**, search `src/components/common/` for a `runai-*` directory that matches the use case before reaching for Quasar.
