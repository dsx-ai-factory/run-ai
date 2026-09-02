# Figma Code Connect - Design System Rules

> Rules for translating Figma designs into production code using the Figma MCP server.

## Technology Stack

- **Framework**: Vue 3.5 (Composition API + Options API)
- **UI Library**: Quasar 2.15 (provides base components: `q-btn`, `q-select`, `q-input`, `q-table`, etc.)
- **Styling**: SCSS with Quasar SCSS variables, scoped styles
- **Language**: TypeScript 5.3 (strict mode)
- **Build**: Vite 5.4
- **Icons**: Phosphor Icons v2 (via `ph`, `ph-fill`, `ph-light`, `ph-thin`, `ph-bold`, `ph-duotone` weight prefixes). Semantic catalog at `src/plugins/runai-icons.catalog.ts`. Custom monochrome glyphs ship through the SVG sprite at `src/assets/icons/runai-glyph/` and render via the `runai-` prefix.

## Token Mapping: Figma to SCSS

All semantic tokens are available as CSS custom properties in `src/assets/scss/vendor/_quasar-variables.scss`. Pre-existing tokens use `--b-` prefix (`--b-text-main`, `--b-text-sub`, `--b-text-label`); all others are unprefixed.

**Figma Foundations file**: `4JHDZ3kV0hkDeIDBLBrfcR` (run-ai Foundations > Color > Semantic Tokens)

### 1. Text


| Figma Token           | CSS Custom Property       | Foundation Ref   | Value                    |
| --------------------- | ------------------------- | ---------------- | ------------------------ |
| text/main             | `--b-text-main`           | `$black-87`      | `rgba(0,0,0,0.87)`       |
| text/sub              | `--b-text-sub`            | `$black-54`      | `rgba(0,0,0,0.54)`       |
| text/tertiary         | `--text-tertiary`         | `$grey-13-5`     | `#939398`                |
| text/inverse          | `--text-inverse`          | `$grey-0`        | `#ffffff`                |
| text/inverse-edit     | `--text-inverse-edit`     | `$white-54`      | `rgba(255,255,255,0.54)` |
| text/inverse-disabled | `--text-inverse-disabled` | `$light-green-2` | `#dcedc8`                |
| text/hint             | `--text-hint`             | `$grey-13-6`     | `#646464`                |
| text/label            | `--b-text-label`          | `$black-60`      | `rgba(0,0,0,0.6)`        |
| text/placeholder      | `--text-placeholder`      | `$black-60`      | `rgba(0,0,0,0.6)`        |
| text/disabled         | `--text-disabled`         | `$black-54`      | `rgba(0,0,0,0.54)`       |
| text/error            | `--text-error`            | `$red-6`         | `#f44336`                |
| text/warning          | `--text-warning`          | `$orange-6`      | `#ff9800`                |
| text/success          | `--text-success`          | `$green-6`       | `#4caf50`                |
| text/brand            | `--text-brand`            | `$nvidia-1`      | `#65a800`                |
| text/link-default     | `--text-link-default`     | `$indigo-14`     | `#304ffe`                |


### 2. Background


| Figma Token          | CSS Custom Property      | Foundation Ref           | Value                   |
| -------------------- | ------------------------ | ------------------------ | ----------------------- |
| bg/primary/default   | `--bg-primary-default`   | `$grey-0`                | `#ffffff`               |
| bg/primary/hover     | `--bg-primary-hover`     | `$grey-3`                | `#eeeeee`               |
| bg/primary/pressed   | `--bg-primary-pressed`   | `$light-green-1`         | `#f1f8e9`               |
| bg/primary/disabled  | `--bg-primary-disabled`  | `$grey-4`                | `#e0e0e0`               |
| bg/primary/info      | `--bg-primary-info`      | `$blue-1`                | `#e3f2fd`               |
| bg/primary/error     | `--bg-primary-error`     | `$red-1`                 | `#ffebee`               |
| bg/primary/warning   | `--bg-primary-warning`   | `$amber-1-5`             | `#ffeecc`               |
| bg/primary/success   | `--bg-primary-success`   | `$green-1`               | `#e8f5e9`               |
| bg/overlay           | `--bg-overlay`           | `$grey-40`               | `rgba(51,51,51,0.4)`    |
| bg/secondary         | `--bg-secondary`         | `$light-grey-40`         | `rgba(241,243,244,0.4)` |
| bg/secondary/default | `--bg-secondary-default` | `$body-background-color` | `#f1f3f4`               |


### 3. Border


| Figma Token    | CSS Custom Property | Foundation Ref   | Value     |
| -------------- | ------------------- | ---------------- | --------- |
| border/default | `--border-default`  | `$grey-5`        | `#bdbdbd` |
| border/strong  | `--border-strong`   | `$grey-7`        | `#757575` |
| border/subtle  | `--border-subtle`   | `$grey-4`        | `#e0e0e0` |
| border/inverse | `--border-inverse`  | `$light-green-5` | `#9ccc65` |
| border/hover   | `--border-hover`    | `$grey-9`        | `#424242` |
| border/pressed | `--border-pressed`  | `$nvidia-1`      | `#65a800` |
| border/error   | `--border-error`    | `$red-5`         | `#ef5350` |
| border/warning | `--border-warning`  | `$amber-14`      | `#ffab00` |


### 4. Interactive

#### 4a. Primary


| Figma Token                      | CSS Custom Property                  | Foundation Ref | Value     |
| -------------------------------- | ------------------------------------ | -------------- | --------- |
| interactive/primary/default/bg   | `--interactive-primary-default-bg`   | `$nvidia-1`    | `#65a800` |
| interactive/primary/default/text | `--interactive-primary-default-text` | `$grey-0`      | `#ffffff` |


#### 4b. Secondary


| Figma Token                                | CSS Custom Property                            | Foundation Ref     | Value                 |
| ------------------------------------------ | ---------------------------------------------- | ------------------ | --------------------- |
| interactive/secondary/default/bg           | `--interactive-secondary-default-bg`           | `$white-0`         | `rgba(255,255,255,0)` |
| interactive/secondary/default/text         | `--interactive-secondary-default-text`         | `$nvidia-1`        | `#65a800`             |
| interactive/secondary/default/icon         | `--interactive-secondary-default-icon`         | `$nvidia-1`        | `#65a800`             |
| interactive/secondary/default/border       | `--interactive-secondary-default-border`       | `$grey-7`          | `#757575`             |
| interactive/secondary/hover/bg             | `--interactive-secondary-hover-bg`             | `$white-0`         | `rgba(255,255,255,0)` |
| interactive/secondary/hover/bg-inverse     | `--interactive-secondary-hover-bg-inverse`     | `$nvidia-1`        | `#65a800`             |
| interactive/secondary/hover/text           | `--interactive-secondary-hover-text`           | `$nvidia-1`        | `#65a800`             |
| interactive/secondary/hover/icon           | `--interactive-secondary-hover-icon`           | `$nvidia-1`        | `#65a800`             |
| interactive/secondary/hover/border         | `--interactive-secondary-hover-border`         | `$grey-4`          | `#e0e0e0`             |
| interactive/secondary/pressed/bg           | `--interactive-secondary-pressed-bg`           | `$grey-4`          | `#e0e0e0`             |
| interactive/secondary/pressed/border       | `--interactive-secondary-pressed-border`       | `$nvidia-1`        | `#65a800`             |
| interactive/secondary/pressed/text         | `--interactive-secondary-pressed-text`         | `$nvidia-1`        | `#65a800`             |
| interactive/secondary/selected/bg          | `--interactive-secondary-selected-bg`          | `$light-green-8`   | `#689f38`             |
| interactive/secondary/selected-disabled/bg | `--interactive-secondary-selected-disabled-bg` | `$light-green-4`   | `#aed581`             |
| interactive/secondary/selected/text        | `--interactive-secondary-selected-text`        | `$grey-0`          | `#ffffff`             |
| interactive/secondary/disabled/bg          | `--interactive-secondary-disabled-bg`          | `$white-0`         | `rgba(255,255,255,0)` |
| interactive/secondary/disabled/border      | `--interactive-secondary-disabled-border`      | `$grey-5`          | `#bdbdbd`             |
| interactive/secondary/disabled/text        | `--interactive-secondary-disabled-text`        | `$light-green-4-5` | `#b2d380`             |
| interactive/secondary/disabled/icon        | `--interactive-secondary-disabled-icon`        | `$light-green-4-5` | `#b2d380`             |


#### 4c. Tertiary


| Figma Token                        | CSS Custom Property                    | Foundation Ref | Value                 |
| ---------------------------------- | -------------------------------------- | -------------- | --------------------- |
| interactive/tertiary/default/bg    | `--interactive-tertiary-default-bg`    | `$white-0`     | `rgba(255,255,255,0)` |
| interactive/tertiary/default/text  | `--interactive-tertiary-default-text`  | `$black-54`    | `rgba(0,0,0,0.54)`    |
| interactive/tertiary/hover/bg      | `--interactive-tertiary-hover-bg`      | `$grey-3`      | `#eeeeee`             |
| interactive/tertiary/hover/text    | `--interactive-tertiary-hover-text`    | `$black-54`    | `rgba(0,0,0,0.54)`    |
| interactive/tertiary/disabled/bg   | `--interactive-tertiary-disabled-bg`   | `$white-0`     | `rgba(255,255,255,0)` |
| interactive/tertiary/disabled/text | `--interactive-tertiary-disabled-text` | `$black-35`    | `rgba(0,0,0,0.35)`    |


#### 4d. Standard


| Figma Token                        | CSS Custom Property                    | Foundation Ref | Value              |
| ---------------------------------- | -------------------------------------- | -------------- | ------------------ |
| interactive/standard/default/bg    | `--interactive-standard-default-bg`    | `$grey-4`      | `#e0e0e0`          |
| interactive/standard/default/text  | `--interactive-standard-default-text`  | `$black-87`    | `rgba(0,0,0,0.87)` |
| interactive/standard/hover/bg      | `--interactive-standard-hover-bg`      | `$grey-5`      | `#bdbdbd`          |
| interactive/standard/hover/text    | `--interactive-standard-hover-text`    | `$black-87`    | `rgba(0,0,0,0.87)` |
| interactive/standard/disabled/bg   | `--interactive-standard-disabled-bg`   | `$grey-3`      | `#eeeeee`          |
| interactive/standard/disabled/text | `--interactive-standard-disabled-text` | `$black-54`    | `rgba(0,0,0,0.54)` |


#### 4e. Naked


| Figma Token                    | CSS Custom Property                | Foundation Ref | Value              |
| ------------------------------ | ---------------------------------- | -------------- | ------------------ |
| interactive/naked/default/bg   | `--interactive-naked-default-bg`   | `$grey-4`      | `#e0e0e0`          |
| interactive/naked/default/text | `--interactive-naked-default-text` | `$black-54`    | `rgba(0,0,0,0.54)` |
| interactive/naked/hover/bg     | `--interactive-naked-hover-bg`     | `$grey-5`      | `#bdbdbd`          |
| interactive/naked/hover/text   | `--interactive-naked-hover-text`   | `$black-54`    | `rgba(0,0,0,0.54)` |


### 5. Navigation

#### 5a. Primary


| Figma Token                       | CSS Custom Property            | Foundation Ref | Value             |
| --------------------------------- | ------------------------------ | -------------- | ----------------- |
| navigation/primary/bg/default     | `--nav-primary-bg-default`     | `$grey-15`     | `#000000`         |
| navigation/primary/bg/selected    | `--nav-primary-bg-selected`    | `$grey-9`      | `#424242`         |
| navigation/primary/icon/default   | `--nav-primary-icon-default`   | `$grey-5`      | `#bdbdbd`         |
| navigation/primary/icon/selected  | `--nav-primary-icon-selected`  | `$grey-1`      | `#fafafa`         |
| navigation/primary/text/default   | `--nav-primary-text-default`   | `$grey-5`      | `#bdbdbd`         |
| navigation/primary/text/selected  | `--nav-primary-text-selected`  | `$grey-1`      | `#fafafa`         |
| navigation/primary/button/default | `--nav-primary-button-default` | `$grey-9`      | `#424242`         |
| navigation/primary/button/hover   | `--nav-primary-button-hover`   | `$grey-6`      | `#9e9e9e`         |
| navigation/primary/button/pressed | `--nav-primary-button-pressed` | `$grey-5`      | `#bdbdbd`         |
| navigation/primary/breadcrumbs    | `--nav-primary-breadcrumbs`    | `$black-40`    | `rgba(0,0,0,0.4)` |


#### 5b. Secondary


| Figma Token                        | CSS Custom Property             | Foundation Ref   | Value     |
| ---------------------------------- | ------------------------------- | ---------------- | --------- |
| navigation/secondary/bg/default    | `--nav-secondary-bg-default`    | `$grey-1`        | `#fafafa` |
| navigation/secondary/bg/selected   | `--nav-secondary-bg-selected`   | `$light-green-2` | `#dcedc8` |
| navigation/secondary/bg/hover      | `--nav-secondary-bg-hover`      | `$grey-4`        | `#e0e0e0` |
| navigation/secondary/icon/default  | `--nav-secondary-icon-default`  | `$grey-9`        | `#424242` |
| navigation/secondary/icon/selected | `--nav-secondary-icon-selected` | `$grey-9`        | `#424242` |
| navigation/secondary/text/default  | `--nav-secondary-text-default`  | `$grey-9`        | `#424242` |
| navigation/secondary/text/selected | `--nav-secondary-text-selected` | `$grey-9`        | `#424242` |


### 6. Stepper


| Figma Token            | CSS Custom Property        | Foundation Ref   | Value     |
| ---------------------- | -------------------------- | ---------------- | --------- |
| stepper/default/icon   | `--stepper-default-icon`   | `$grey-6`        | `#9e9e9e` |
| stepper/default/text   | `--stepper-default-text`   | `$grey-6`        | `#9e9e9e` |
| stepper/hover/bg       | `--stepper-hover-bg`       | `$grey-4`        | `#e0e0e0` |
| stepper/hover/icon     | `--stepper-hover-icon`     | `$grey-6`        | `#9e9e9e` |
| stepper/hover/text     | `--stepper-hover-text`     | `$grey-6`        | `#9e9e9e` |
| stepper/selected/icon  | `--stepper-selected-icon`  | `$light-green-8` | `#689f38` |
| stepper/completed/icon | `--stepper-completed-icon` | `$light-green-8` | `#689f38` |
| stepper/completed/text | `--stepper-completed-text` | `$grey-14-5`     | `#4a4a4a` |
| stepper/disabled/icon  | `--stepper-disabled-icon`  | `$grey-12-5`     | `#c7c7c7` |
| stepper/disabled/text  | `--stepper-disabled-text`  | `$grey-12-5`     | `#c7c7c7` |
| stepper/error/icon     | `--stepper-error-icon`     | `$red-5`         | `#ef5350` |
| stepper/error/text     | `--stepper-error-text`     | `$red-5`         | `#ef5350` |


### 7. Icon


| Figma Token           | CSS Custom Property       | Foundation Ref     | Value              |
| --------------------- | ------------------------- | ------------------ | ------------------ |
| icon/default          | `--icon-default`          | `$black-54`        | `rgba(0,0,0,0.54)` |
| icon/interactive      | `--icon-interactive`      | `$grey-8`          | `#616161`          |
| icon/tertiary         | `--icon-tertiary`         | `$grey-13-5`       | `#939398`          |
| icon/inverse          | `--icon-inverse`          | `$grey-1`          | `#fafafa`          |
| icon/inverse-disabled | `--icon-inverse-disabled` | `$light-green-3-5` | `#c3dda8`          |
| icon/disabled         | `--icon-disabled`         | `$black-35`        | `rgba(0,0,0,0.35)` |
| icon/hover            | `--icon-hover`            | `$grey-1`          | `#fafafa`          |
| icon/error            | `--icon-error`            | `$red-6`           | `#f44336`          |
| icon/warning          | `--icon-warning`          | `$amber-14`        | `#ffab00`          |
| icon/info             | `--icon-info`             | `$blue-6`          | `#2196f3`          |


### 8. Feedback


| Figma Token         | CSS Custom Property     | Alias               | Foundation Ref | Value     |
| ------------------- | ----------------------- | ------------------- | -------------- | --------- |
| feedback/default/bg | `--feedback-default-bg` | `--b-feedback-default-bg` | `$grey-14` | `#616161` |
| feedback/text       | `--feedback-text`       | `--b-feedback-text`       | `$grey-0`  | `#ffffff` |
| feedback/success/bg | `--feedback-success`    | `--b-feedback-success-bg` | `$green-6` | `#4caf50` |
| feedback/warning/bg | `--feedback-warning`    | `--b-feedback-warning-bg` | `$amber-14`| `#ffab00` |
| feedback/error/bg   | `--feedback-error`      | `--b-feedback-error-bg`   | `$red-5`   | `#ef5350` |
| feedback/info/bg    | `--feedback-info`       | `--b-feedback-info-bg`    | `$blue-6`  | `#2196f3` |

> `--b-feedback-*` is a naming alias used in tickets/docs — in code always use the `--feedback-*` property.


### 9. Table


| Figma Token    | CSS Custom Property | Foundation Ref | Value     |
| -------------- | ------------------- | -------------- | --------- |
| table/bg       | `--table-bg`        | `$grey-0`      | `#ffffff` |
| table/hover    | `--table-hover`     | `$grey-2`      | `#f5f5f5` |
| table/selected | `--table-selected`  | `$grey-10-5`   | `#f0f0f0` |
| table/disabled | `--table-disabled`  | `$grey-1`      | `#fafafa` |


### 10. Decorative


| Figma Token      | CSS Custom Property | Foundation Ref | Value     |
| ---------------- | ------------------- | -------------- | --------- |
| decorative/dark  | `--decorative-dark` | `$grey-15`     | `#000000` |


### 11. Charts


| Figma Token        | CSS Custom Property    | Foundation Ref   | Value     |
| ------------------ | ---------------------- | ---------------- | --------- |
| charts/blue        | `--charts-blue`        | `$decorative-1`  | `#2caffe` |
| charts/orange      | `--charts-orange`      | `$decorative-2`  | `#fe6a35` |
| charts/lime        | `--charts-lime`        | `$decorative-3`  | `#00e272` |
| charts/purple      | `--charts-purple`      | `$decorative-4`  | `#544fc5` |
| charts/light-green | `--charts-light-green` | `$decorative-5`  | `#91e8e1` |
| charts/pink        | `--charts-pink`        | `$decorative-6`  | `#e560ff` |
| charts/mustard     | `--charts-mustard`     | `$decorative-7`  | `#ffb15a` |
| charts/lavender    | `--charts-lavender`    | `$decorative-8`  | `#6b8abc` |
| charts/cyan        | `--charts-cyan`        | `$decorative-9`  | `#00e4ca` |
| charts/grey        | `--charts-grey`        | `$decorative-10` | `#cccccc` |


### Spacing

T-shirt size scale. Available as CSS custom properties and utility classes.


| Figma Token | CSS Custom Property | Utility Class (example) | Value  |
| ----------- | ------------------- | ----------------------- | ------ |
| space/3xs   | `--space-3xs`       | `b-pa-3xs`, `b-gap-3xs` | `3px`  |
| space/2xs   | `--space-2xs`       | `b-pa-2xs`, `b-gap-2xs` | `5px`  |
| space/xs    | `--space-xs`        | `b-pa-xs`, `b-gap-xs`   | `7px`  |
| space/s     | `--space-s`         | `b-pa-s`, `b-gap-s`     | `10px` |
| space/m     | `--space-m`         | `b-pa-m`, `b-gap-m`     | `16px` |
| space/l     | `--space-l`         | `b-pa-l`, `b-gap-l`     | `20px` |
| space/xl    | `--space-xl`        | `b-pa-xl`, `b-gap-xl`   | `26px` |
| space/2xl   | `--space-2xl`       | `b-pa-2xl`, `b-gap-2xl` | `40px` |


**Utility class pattern**: `b-{property}-{size}` where property is:

- Padding: `pa` (all), `pt/pr/pb/pl` (sides), `px/py` (axes)
- Margin: `ma` (all), `mt/mr/mb/ml` (sides), `mx/my` (axes)
- Gap: `gap`

### Typography

| Figma Token | Font | Size | Weight | Line Height |
|---|---|---|---|---|
| `text-b-h1` | Roboto | 45px | 700 (Bold) | 100% (45px) |
| `text-b-h6` | Roboto | 16px | 400 (Regular) | 22.5px |
| `text-b-h6-bold` | Roboto | 16px | 700 (Bold) | 22.5px |
| `text-b-h6-md` | Roboto | 16px | 500 (Medium) | 22.5px |
| `text-b-subtitle1` | Roboto | 14px | 500 (Medium) | 20px |
| `text-b-subtitle2` | Roboto | 14px | 400 (Regular) | 19.5px |
| `text-b-body1` | Roboto | 12px | 400 (Regular) | 16.8px |
| `text-b-body2` | Roboto | 10px | 400 (Regular) | 14px |
| `text-b-mono` | Roboto Mono | 12px | 400 (Regular) | 1.3 |

### Border Radius

| Figma Token | CSS Custom Property | Value |
|---|---|---|
| `radius-0` | `--b-border-radius-radius-0` | `0px` |
| `radius-xs` | `--b-border-radius-radius-xs` | `3px` |
| `radius-s` | `--b-border-radius-radius-s` | `5px` |
| `radius-m` | `--b-border-radius-radius-m` | `10px` |
| `radius-l` | `--b-border-radius-radius-l` | `16px` |
| `radius-full` | `--b-border-radius-radius-full` | `9999px` |

### Shadows


| Figma Token | CSS Custom Property | Utility Class | `box-shadow` Value                     |
| ----------- | ------------------- | ------------- | -------------------------------------- |
| shadow/s    | `--shadow-s`        | `b-shadow-s`  | `-1px 0px 7px 0px rgba(0, 0, 0, 0.15)` |
| shadow/m    | `--shadow-m`        | `b-shadow-m`  | `0px 4px 12px 0px rgba(0, 0, 0, 0.20)` |
| shadow/l    | `--shadow-l`        | `b-shadow-l`  | `0px 2px 6px 25px rgba(0, 0, 0, 0.26)` |


## Implementation Rules (Figma → Code)

These rules prevent common mistakes when translating a Figma design into code.

### Rule 1: Pixel-perfect fidelity — only render what's in the Figma

- **Do NOT add visual elements** (separators, borders, dividers, icons, backgrounds) that are not explicitly visible in the Figma screenshot.
- **Do NOT remove visual elements** that are in the Figma screenshot.
- If the Figma shows no separators between list items — don't add `<q-separator>`.
- If the Figma shows no border on a container — don't add one.
- Always cross-reference the Figma screenshot before adding any decorative element.

### Rule 2: Use typography classes — never manually specify font properties

- Get the text style name from the Figma MCP design context output (e.g. `text-b-subtitle2`).
- **Always use the corresponding typography class** (e.g. `class="text-b-subtitle2"`) from `src/assets/scss/base/_typography.scss`. These classes already include `font-size`, `font-weight`, `line-height`, `color`, and `letter-spacing`.
- **Do NOT manually specify** `font-size`, `font-weight`, `line-height`, `letter-spacing`, or `color` in SCSS when a typography class covers them. Only add SCSS overrides for properties that differ from the class defaults (e.g. `color: var(--b-text-sub)` for secondary text, `color: var(--text-disabled)` for disabled text).
- Common mistake: assuming titles are bold (700) when the Figma style says Regular (400). The size contrast between title and caption can create a visual impression of boldness — always check the actual weight value.

### Rule 3: Avoid `!important` — use it only when Quasar forces it

**Default**: Do NOT use `!important`. Use higher specificity selectors instead (e.g. `.my-class.q-menu` instead of `.my-class`).

**Exception**: Quasar uses `!important` on a few internal styles. These can **only** be overridden with `!important` (specificity alone won't work). When you must use `!important`, always add an inline comment explaining why:


| Quasar selector  | Property with `!important`      | What to do                                                                                                    |
| ---------------- | ------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| `.q-item__label` | `line-height: 1.2em !important` | Use `line-height: <value> !important; // overrides Quasar's .q-item__label { line-height: 1.2em !important }` |
| `.q-menu`        | `position: fixed !important`    | Don't override position                                                                                       |


### Rule 4: Always reset Quasar defaults on popup menus (`q-menu`)

When styling a `q-menu` dropdown/popup, use `.my-class.q-menu` for higher specificity (no `!important` needed — Quasar's `background`, `box-shadow`, and `border` do NOT use `!important`):

```scss
.my-menu-class.q-menu {
  background: var(--bg-primary-default);
  box-shadow: var(--shadow-m);  // or shadow-s
  border: none;                 // removes phantom borders
  border-radius: var(--space-2xs);
}
```

Quasar's default `.q-menu` has a multi-layered `box-shadow` that can look like a border. Always override `box-shadow` and add `border: none`.

### Rule 5: Use semantic tokens for ALL colors — never use hardcoded rgba/hex

- Background colors: use `--bg-*` tokens (e.g. `--bg-primary-default`, not `#ffffff`)
- Text colors: use `--b-text-*` or `--text-*` tokens (e.g. `--b-text-main`, not `rgba(0,0,0,0.87)`)
- Hover backgrounds: use `--bg-primary-hover`, not `rgba(0,0,0,0.08)`
- Border/separator colors: use `--border-subtle`, not `rgba(0,0,0,0.12)`
- Do NOT use `--interactive-primary-*` tokens (green/brand colors) for standard dropdown menus or content backgrounds — those are reserved for primary action buttons.

### Rule 6: Menus render in a portal — styles must be unscoped

`q-menu` renders outside the component DOM tree (in a portal). Styles targeting the menu content **must NOT** use `<style scoped>` — use plain `<style lang="scss">` so selectors can reach the portal-rendered elements.

## Translation Rules

### 1. React + Tailwind to Vue 3 + Quasar + SCSS

Figma MCP generates React + Tailwind code. Always translate:

```
// Figma output (DON'T use this)
<div className="flex items-center gap-4 p-4">

// Translated (USE this)
<div class="row items-center q-gutter-sm q-pa-md">
```

Use Quasar utility classes:

- `row` / `column` instead of `flex` / `flex-col`
- `items-center`, `justify-between` (same as Tailwind)
- `q-pa-{size}`, `q-ma-{size}`, `q-gutter-{size}` for spacing
- `q-mr-sm`, `q-ml-md`, etc. for individual margins

### 2. Component Reuse

ALWAYS check for an existing `runai-*` component before creating new ones.
Components live in `src/components/common/runai-*/`.

Key components:

- `runai-select` - Full select input (wraps `q-select`)
- `runai-select-menu` - Dropdown menu popup (wraps `q-menu` + `q-list`)
- `runai-btn` - Button component
- `runai-input` - Text input
- `runai-tooltip` - Icon with tooltip popup (wraps `q-btn` + `q-tooltip`)
- `runai-table` - Data table
- `runai-base-dialog` - Modal dialog
- `runai-side-drawer` - Side panel
- `runai-svg-icon` - SVG icon component

#### Figma component → RunAI component map

**Figma Code Connect is now published** on ~40 components in the 🟩 run:ai Components library, so `get_design_context` / `get_code_connect_map` will usually return the correct `runai-*` source directly — **prefer that output**. This table is the fallback: use it when a node is *not yet* Code-Connected (e.g. components not built in code, product-file-local names that diverge from the library, or components added since this list).

When you do use it, render the mapped component using the **exact import path in that row's "Import barrel" column** — most are under `@/components/common/`, but some live in `@/components/infra/…`, nested paths (e.g. `runai-table/runai-table-no-data`), or have no `runai-` prefix (`notification-alert-bar`). Don't assume a fixed barrel. Keep it in sync as the design system aligns.

| Figma component name | RunAI component | Import barrel |
|---|---|---|
| `Table_Workloads`, `Table_*` (page grid) | `runai-table-wrapper` | `@/components/common/runai-table-wrapper` |
| `Drawer` | `runai-side-drawer` | `@/components/common/runai-side-drawer` |
| `Drawer Body` | `runai-drawer-content` | `@/components/common/runai-drawer-content` |
| `Dialog`, `DialogGeneric` | `runai-base-dialog` | `@/components/common/runai-base-dialog` |
| `feedback/dialog-preset` (confirmation/delete) | `runai-delete-modal` | `@/components/common/runai-delete-modal` (verify; `runai-base-dialog` for generic) |
| `menu/select-menu` | `runai-select-menu` | `@/components/common/runai-select-menu` |
| `Breadcrumbs Item` | `runai-breadcrumbs` | `@/components/common/runai-breadcrumbs` |
| `Brand Logo`, `brand/logo-nvidia` | `runai-brand-logo` | `@/components/common/runai-brand-logo` |
| `Icon` | `runai-svg-icon` | `@/components/common/runai-svg-icon` |
| `Tooltip Button`, `Tooltip`, `Tooltip Content` | `runai-tooltip` | `@/components/common/runai-tooltip` |
| `display/tooltip-data` | `runai-data-tooltip` | `@/components/common/runai-data-tooltip` |
| `form/select-input-field` | `runai-select` | `@/components/common/runai-select` |
| `form/toggle` | `runai-toggle` | `@/components/common/runai-toggle` |
| `form/textarea-input` | **current:** `runai-description-input` · **planned:** `runai-textarea` | `@/components/common/runai-description-input` |
| `form/form-expansion-item` | `runai-expansion-item` | `@/components/common/runai-expansion-item` |
| `date-picker/calendar`, `date-picker/datepicker-with-predefined` | `runai-datepicker` / `runai-datepicker-with-predefined` | `@/components/common/runai-datepicker` |
| `action/button` (canonical set: Primary/Secondary/Tertiary/Standard), `action/action-button`, `action/button-action-bar`, `action/button-label`, `action/icon-button`, `Tertiary Default Button`, `Secondary`, `Secondary Default Button`, `Standard` | `runai-btn` | `@/components/common/runai-btn` |
| `Badge`, `display/badge-mono` | `runai-badge-mono` | `@/components/common/runai-badge-mono` |
| `Card header` | `runai-form-card-section` | `@/components/common/runai-form-card-section` (verify; `runai-section` for non-form) |
| `In-form dialog footer` | `runai-form-footer` | `@/components/common/runai-form-footer` (verify props) |
| `Drawer_Search Bar` | `runai-general-search` | `@/components/common/runai-general-search` (verify props) |
| `navigation/tab-group`, `Tab group`, `tabs_single-tab` | `runai-tab` | `@/components/common/runai-tab` |
| `display/log-editor` | `runai-log-reader` | `@/components/common/runai-log-reader` |
| `display/code-editor` | `runai-editor` (fullscreen → `runai-editor-modal`) | `@/components/common/runai-editor` |
| `Chip` | `runai-filter-chip` | `@/components/common/runai-filter-chip` (verify props) |
| `entity/filter`, `Table_Filter_Dialog` | `runai-advanced-filters` | `@/components/common/runai-advanced-filters` (verify props) |
| `display/empty-state` | `runai-table-no-data` (grid) / `runai-drawer-no-data` (drawer) | `@/components/common/runai-table/runai-table-no-data` |
| `display/notification-alert-bar` | `notification-alert-bar` | `@/components/common/notification-alert-bar` (no `runai-` prefix) |
| `navigation/pagination-row` | *inside* `runai-table` (Quasar `q-table` built-in — no standalone component) | `@/components/common/runai-table` |
| `form/radio`, `form/radio-group` | `runai-radio-options` | `@/components/common/runai-radio-options` |
| `Duration Input Section` | `runai-duration-input` | `@/components/common/runai-duration-input` |
| `action/copy-box` | `runai-copyable-code-block` / `runai-copy-to-clipboard-btn` | `@/components/common/runai-copyable-code-block` |
| `action-bar` (Action Bar) | `runai-page-actions` | `@/components/common/runai-page-actions` |
| `form/checkbox` | `runai-checkbox-list` | `@/components/common/runai-checkbox-list` |
| `Left Panel Stepper` | `runai-stepper` | `@/components/common/runai-stepper` |
| `form/input-text-field`, `Input with label` | `runai-input` | `@/components/common/runai-input` |
| `display/folding-sub-section` | `runai-folding-sub-section` | `@/components/common/runai-folding-sub-section` |
| `section-title` | `runai-section-title` | `@/components/common/runai-section-title` |
| `form/scope-selector-input` | `runai-scope-section` (multi → `runai-multi-scope-input`) | `@/components/common/runai-scope-section` |
| `display/org-tree` | `runai-org-tree` (in dialog → `runai-org-tree-with-popover`) | `@/components/common/runai-org-tree` |
| `navigation/header-global` | `runai-header` | `@/components/infra/runai-layout/runai-header` |
| `navigation/Toolbar` (sidebar nav) | `runai-side-nav` (+ `runai-side-nav-menu-item`, `runai-sub-menu`, `runai-sub-menu-link`) | `@/components/infra/runai-layout/runai-side-nav` |

> Rows below the filters row and several above are **confirmed by the "Design System Alignment" tickets** — each ticket carries the authoritative Figma node + code path, so these mappings are verified, not heuristic.

**Open / in-progress align tickets (mapping known, implementation pending — use these targets, expect drift until the ticket closes):**

| Figma component | RunAI target | Status |
|---|---|---|
| `form/textarea-input` | `runai-textarea` (new shared component; today `runai-description-input` / inline `q-input autogrow`) | Planned — becomes the standard target for the `form/textarea-input` row above once built |
| interactive `q-icon` (clickable + tooltip) | `runai-icon-button` (new; **not** for decorative icons — those stay `q-icon`) | Planned |
| `display/avatar` | `runai-avatar` (new shared; today ad-hoc in profile-dropdown + user-settings) | Planned |
| `form/load-asset-bar` | composite in `hybrid-section-wrapper.vue` `#content-header` slot (no standalone) | Planned |
| `display/minimap`, `action/view-controller` | Vue Flow `@vue-flow/minimap` / `@vue-flow/controls` (3rd-party, style overrides only) in `ai-application-diagram.vue` | Planned |
| Leave-this-page modal | `runai-base-dialog` + unsaved-changes guard (`runai-form-wrapper`) | Planned |

Notes:
- Figma components prefixed `✖️` are **deprecated** — do not map or reproduce them; use the current `runai-*` equivalent and flag the drift.
- `*_Workloads` / `Action_Bar_*` / `Header_*` / `Navigation` and similar are **product composites** (page assemblies), not reusable components — build them from the primitives above, not as a single component.
- `asset/*`, `object/*`, `nav/*`, `action/<verb>` nodes are individual icons → `runai-svg-icon` with the matching icon name, not standalone components.
- **No current wrapper** (build with Quasar / utilities and flag for alignment): `Toast` → Quasar `Notify` / notification util (alignment still open); `display/tooltip-table` → not yet a shared component; `chart/*` → `runai-resources-chart` only for resource charts, otherwise Highcharts wrapper.
- Coverage note: this map was harvested from the **Workloads**, **User Settings**, **Projects** (all V3), and **General Settings** (V2) design files, the run:ai Components library, and the **Design System Alignment tickets** (Done + open). Curated, not exhaustive — it covers the primitives surfaced from those sources so far; new components may need adding as the design system evolves. `runai-numeric-input` was removed (unused — no Figma component, no code references). Extend the table as new components are added or align tickets close.
- Naming caveat: **V3** files (e.g. Workloads) use the library `category/name` convention; **V2** files (e.g. General Settings) use looser, ad-hoc names (`Secondary`, `Standard`, `Badge`, `Card header`). Match V2 names by concept, not convention, and prefer the V3 library name when both exist.

### 3. Icons

Phosphor is the canonical icon font. Always derive icon names from `runai-icons.catalog.ts` so semantic intent stays in sync with Figma.

```vue
<!-- Catalog-driven (preferred) -->
<q-icon :name="toQuasarIconName(runaiIcons.action.delete)" />

<!-- Direct Phosphor when not yet in the catalog -->
<q-icon name="ph ph-trash" />
<q-icon name="ph-fill ph-eye" />

<!-- Custom monochrome glyph from the runai-glyph SVG sprite -->
<q-icon name="runai-access" />

<!-- Multicolor / illustrative SVG (not part of the icon font) -->
<runai-svg-icon name="icon-name" size="20px" />
```

Weight prefixes: omit (`ph ph-X` = regular) or use `ph-thin`, `ph-light`, `ph-bold`, `ph-fill`, `ph-duotone`. FontAwesome (`fa-*`) is removed — never reintroduce it.

#### Icon sizing (Figma `action/icon-button`, Components node `123:2855`)

Two CSS custom properties define the canonical sizes (`src/assets/scss/abstracts/_icon-sizes.scss`):

| Token | Value | Where |
|---|---|---|
| `--icon-size-default` | **20px** | Nav subjects, primary CTAs, search boxes, action buttons, page-level toolbar icons |
| `--icon-size-small` | **16px** | Table cells, scope tree nodes, alert bars, status badges, header dropdown rows |

Three icon-button variants (icon-only `q-btn`). Apply the helper class to the `q-btn`:

| Helper class | Button frame | Icon size | Use case |
|---|---|---|---|
| `.runai-icon-btn--default` | 38×38 | 20px | Primary toolbars, page actions, bulk-action row |
| `.runai-icon-btn--small` | 32×32 | 16px | Dropdown rows, dense toolbars, scope picker trigger |
| `.runai-icon-btn--xsmall` | 24×24 | 16px | Table row actions, tooltip triggers, close/dismiss buttons |

Standalone `q-icon` (no button wrapper): use `size="20px"` for nav / CTAs, `size="16px"` for table / tree / alert contexts. Never use `14px`, `12px`, `10px`, `xs`, or `sm` for icon sizing.

### 4. Common Interface

Most select/dropdown components use `ISelectOption`:

```typescript
interface ISelectOption {
  label: string;
  value: string | boolean | number | null | undefined;
  icon?: string;
  disable?: boolean;
  toolTip?: string;
  disabledTooltip?: string;
  svg?: string;
  subLabel?: string;
  subLabelColor?: string;
}
```

### 5. Styling Approach

- Use scoped SCSS: `<style lang="scss" scoped>`
- Reference SCSS variables from `src/assets/scss/vendor/_quasar-variables.scss`
- Use CSS custom properties for dynamic/brandable values
- Avoid inline styles unless for dynamic values
- Use semantic token CSS custom properties for all design-system colors (see Token Mapping section)

### 6. File Structure

New components follow this pattern:

```
src/components/common/runai-{name}/
  runai-{name}.vue     # Component file
  index.ts             # Re-export: export { default as RunaiName } from './runai-{name}.vue'
```

## Figma File Reference

- **Foundations file**: `4JHDZ3kV0hkDeIDBLBrfcR` (run-ai Foundations)
- **Components file**: `u76OCKmVoZnoA3T18yupbg` (run-ai Components)
- **File URL**: https://www.figma.com/design/u76OCKmVoZnoA3T18yupbg/run-ai-Components

### Shared design libraries (for `search_design_system`)

The org's published libraries, discoverable via `get_libraries` on any product file. Pass a `libraryKey` as `includeLibraryKeys` to scope a `search_design_system` query.

- **🟩 run:ai Components** (the component library): `lk-eb165bc7ecbfb5030c098ef1036c1c3c26d9c211cfe8f38ac399e768d4a501669e80e3be4afc5c2695f796f3b1b02cb08050398dc88e8553e3f4a26bebeda40a`
- **🟩 run:ai Foundations** (tokens/styles): `lk-7f4ac420b19197bc83e326d323e0ece2fafc3b7ba8a69bccb2ea5960773d057f18d98caf03e11981bd3d460a59aa226fafb49d13eff6aac768652b5e0d0194a9`

Library component naming convention: `category/name` (e.g. `form/toggle`, `action/button`, `date-picker/calendar`). A `✖️.` prefix marks a **deprecated** component — never map or reproduce it.

### Resolving a Figma component not in the map above

The mapping table is a curated subset, not exhaustive (the library has 400+ components). When `get_design_context` returns a component name not listed:

1. **Match by category/name to a `runai-*`** — the convention is consistent (`form/*` → form inputs, `action/button*` → `runai-btn`, `display/*` → display widgets, `asset/*`/`object/*`/`nav/*` → icons via `runai-svg-icon`).
2. **Confirm the code component exists** — search `src/components/` (usually `common/`, sometimes `infra/`) for the component whose name matches the concept (see the Quasar→RunAI map in `use-runai-components`). Use the exact path you find, not an assumed `common/` barrel.
3. **If still ambiguous**, call `search_design_system` scoped to the run:ai Components `libraryKey` to read the component's description (they include variant + usage guidance), then pick the closest `runai-*`.
4. **If no `runai-*` exists**, build with Quasar per the rules above and flag the gap for alignment.
5. When you resolve a new mapping with confidence, **add a row to the table above** so it compounds.
