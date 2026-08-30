// Quasar `iconMapFn` — maps an app icon name to a Quasar render config.
//
// This is the glue that lets the codebase use two icon systems through a single
// `q-icon`/icon name:
//   - Phosphor icons: names like `ph ph-funnel` or `ph-bold ph-cube` (a CSS class
//     applied to an <i>). Returned as `{ cls }`.
//   - Custom Run.ai glyphs: names like `runai-cluster` or `runai-outlined-access-rule`,
//     rendered from the SVG sprite (registered by vite-plugin-svg-icons as
//     `<symbol id="runai-<name>">`) via `<use href="#runai-<name>">`. Returned as the
//     Quasar `svguse:` form. q-icon emits `<svg><use ...>`, which inherits `color`
//     because the sprite normalizes fills to `currentColor`.
//   - Anything else: returns `undefined`, so Quasar falls back to its default handling
//     (e.g. the active `iconSet`).

export type RunaiIconMapEntry = { cls: string } | { icon: string };

export function iconMapFn(iconName: string): RunaiIconMapEntry | undefined {
  if (iconName.startsWith("ph")) {
    return { cls: iconName };
  }
  if (iconName.startsWith("runai-")) {
    return { icon: `svguse:#${iconName}` };
  }
  return undefined;
}
