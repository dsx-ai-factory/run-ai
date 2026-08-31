/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export type PhosphorWeight = "thin" | "light" | "regular" | "bold" | "fill" | "duotone";
export type RunaiGlyphVariant = "outlined" | "filled";

export interface IIconEntry {
  phosphor: string | null;
  weight: PhosphorWeight;
  custom?: { glyph: string; variant?: RunaiGlyphVariant };
}

export const runaiIcons = {
  nav: {
    // Side-nav primary subjects. Mirrors the Figma "Foundations → Navigation" row exactly.
    // Variant glyphs are exposed separately so existing flat glyph names remain stable.
    analytics: { phosphor: "chart-line", weight: "bold" },
    organization: { phosphor: "network", weight: "bold" },
    access: { phosphor: null, weight: "bold", custom: { glyph: "access" } },
    accessVariants: {
      outlined: { phosphor: null, weight: "bold", custom: { glyph: "access", variant: "outlined" } },
      filled: { phosphor: null, weight: "bold", custom: { glyph: "access", variant: "filled" } },
    },
    workloadsManager: {
      outlined: { phosphor: null, weight: "bold", custom: { glyph: "workloads-manager", variant: "outlined" } },
      filled: { phosphor: null, weight: "bold", custom: { glyph: "workloads-manager", variant: "filled" } },
    },
    policy: { phosphor: "clipboard-text", weight: "bold" },
    admin: { phosphor: "gear-six", weight: "bold" },
    chevron: {
      left: { phosphor: "caret-left", weight: "bold" },
      right: { phosphor: "caret-right", weight: "bold" },
      up: { phosphor: "caret-up", weight: "bold" },
      down: { phosphor: "caret-down", weight: "bold" },
    },
    doubleChevron: {
      left: { phosphor: "caret-double-left", weight: "bold" },
      right: { phosphor: "caret-double-right", weight: "bold" },
    },
    // "Templates" entry inside the setup dropdown — Phosphor's grid-nine matches
    // the fa-game-board-simple visual UX picked in Figma. Not in the canonical Foundations
    // navigation row (it's a sub-item of the setup dropdown, not a top-level side-nav).
    templates: { phosphor: "grid-nine", weight: "bold" },
  },
  action: {
    add: { phosphor: "plus", weight: "bold" },
    remove: { phosphor: "minus", weight: "bold" },
    edit: { phosphor: "pencil-simple", weight: "fill" },
    copy: { phosphor: "copy", weight: "fill" },
    delete: { phosphor: "trash", weight: "fill" },
    // Renamed from `drag` per Figma (action/grab).
    grab: { phosphor: "dots-six-vertical", weight: "bold" },
    move: { phosphor: "arrows-down-up", weight: "bold" },
    show: { phosphor: "eye", weight: "fill" },
    hide: { phosphor: "eye-slash", weight: "fill" },
    lock: { phosphor: "lock", weight: "fill" },
    unlock: { phosphor: "lock-open", weight: "fill" },
    favorite: { phosphor: "star", weight: "fill" },
    review: { phosphor: "file", weight: "fill" },
    details: { phosphor: "file-text", weight: "fill" },
    run: { phosphor: "play", weight: "fill" },
    stop: { phosphor: "stop", weight: "fill" },
    refresh: { phosphor: "arrow-clockwise", weight: "bold" },
    check: { phosphor: "check", weight: "bold" },
    cancelError: { phosphor: "x-circle", weight: "fill" },
    close: { phosphor: "x", weight: "bold" },
    menu: { phosphor: "dots-three-outline-vertical", weight: "fill" },
    // Renamed from `expand` per Figma (action/maximize).
    maximize: { phosphor: "arrows-out", weight: "bold" },
    // Renamed from `collapse` per Figma (action/minimize).
    minimize: { phosphor: "arrows-in", weight: "bold" },
    // Renamed from `unfold` per Figma (action/expand-all).
    expandAll: { phosphor: "arrows-out-line-vertical", weight: "bold" },
    // Renamed from `fold` per Figma (action/collapse-all).
    collapseAll: { phosphor: "arrows-in-line-vertical", weight: "bold" },
    expandView: { phosphor: "picture-in-picture", weight: "bold" },
    back: { phosphor: "arrow-left", weight: "bold" },
    externalLink: { phosphor: "arrow-square-out", weight: "bold" },
    download: { phosphor: "download", weight: "fill" },
    upload: { phosphor: "upload", weight: "bold" },
    loadAsset: { phosphor: "download-simple", weight: "bold" },
    search: { phosphor: "magnifying-glass", weight: "bold" },
    zoomIn: { phosphor: "magnifying-glass-plus", weight: "bold" },
    zoomOut: { phosphor: "magnifying-glass-minus", weight: "bold" },
    reset: { phosphor: "arrow-counter-clockwise", weight: "bold" },
    sortDown: { phosphor: "arrow-down", weight: "bold" },
    sortUp: { phosphor: "arrow-up", weight: "bold" },
    logs: { phosphor: "list-dashes", weight: "bold" },
    terminal: { phosphor: "terminal-window", weight: "bold" },
    // Renamed from `errorData` per Figma (action/data-unavailable).
    dataUnavailable: { phosphor: "file-text", weight: "bold" },
    // New: long-arrow indicator inside SAML/OIDC mappers (was fa-arrow-right-long catalog gap).
    forward: { phosphor: "arrow-right", weight: "bold" },
    // New: "apply parent values" button in workload creation (was fa-arrow-down-to-square gap).
    assignDown: { phosphor: "archive-tray", weight: "bold" },
  },
  editor: {
    columns: { phosphor: "columns", weight: "bold" },
    // Renamed from `frame` per Figma (editor/zoom-to-fit).
    // Per Yuval's correction this is the catalog home for fa-expand / fa-arrows-maximize / fa-maximize.
    zoomToFit: { phosphor: "corners-out", weight: "bold" },
  },
  status: {
    success: { phosphor: "check-circle", weight: "bold" },
    error: { phosphor: "warning-circle", weight: "bold" },
    pending: { phosphor: "hourglass", weight: "bold" },
    warning: { phosphor: "warning", weight: "bold" },
    loading: { phosphor: "circle-notch", weight: "bold" },
    // New from catalog gaps:
    connectivity: { phosphor: "wifi-high", weight: "bold" },
    debug: { phosphor: "bug", weight: "bold" },
    verified: { phosphor: "shield-check", weight: "bold" },
  },
  object: {
    info: { phosphor: "info", weight: "bold" },
    help: { phosphor: "question", weight: "bold" },
    scope: { phosphor: "tree-view", weight: "fill" },
    userLocked: { phosphor: "user-focus", weight: "fill" },
    secret: { phosphor: "password", weight: "bold" },
    // Figma "object/instructions": 3 horizontal lines with folded bottom-right corner.
    // ph-note matches exactly (note has the same 3-line + folded-corner shape).
    // ph-receipt was wrong — receipt has a zigzag bottom, not a folded corner.
    instructions: { phosphor: "note", weight: "bold" },
    calendar: { phosphor: "calendar-dot", weight: "bold" },
    topology: { phosphor: null, weight: "bold", custom: { glyph: "topology" } },
    topologyVariants: {
      outlined: { phosphor: null, weight: "bold", custom: { glyph: "topology", variant: "outlined" } },
      filled: { phosphor: null, weight: "bold", custom: { glyph: "topology", variant: "filled" } },
    },
    // New from catalog gaps:
    connection: { phosphor: "plug", weight: "bold" },
    global: { phosphor: "globe", weight: "bold" },
    secureFile: { phosphor: "file-lock", weight: "bold" },
    folder: { phosphor: "folder-simple", weight: "bold" },
    folderTree: { phosphor: "tree-structure", weight: "bold" },
    linked: { phosphor: "link", weight: "bold" },
    unlinked: { phosphor: "link-break", weight: "bold" },
    accessKey: { phosphor: "lock", weight: "bold" },
  },
  entity: {
    dataSource: { phosphor: "database", weight: "bold" },
    filter: { phosphor: "funnel", weight: "bold" },
    template: { phosphor: "certificate", weight: "bold" },
    deletedWorkload: { phosphor: "trash-simple", weight: "bold" },
    node: { phosphor: "cube", weight: "bold" },
    // Node pools reuse the Nodes glyph (cube) per UX — Phosphor has no "cubes" (confirmed via Figma Foundations node 2818-403).
    nodePool: { phosphor: "cube", weight: "bold" },
    rule: { phosphor: null, weight: "bold", custom: { glyph: "rule" } },
    ruleVariants: {
      outlined: { phosphor: null, weight: "bold", custom: { glyph: "rule", variant: "outlined" } },
      filled: { phosphor: null, weight: "bold", custom: { glyph: "rule", variant: "filled" } },
    },
    accessRule: {
      outlined: { phosphor: null, weight: "bold", custom: { glyph: "access-rule", variant: "outlined" } },
      filled: { phosphor: null, weight: "bold", custom: { glyph: "access-rule", variant: "filled" } },
    },
    project: { phosphor: "suitcase", weight: "bold" },
    environment: { phosphor: "x-square", weight: "bold" },
    resource: { phosphor: "hard-drives", weight: "bold" },
    credential: { phosphor: "key", weight: "bold" },
    policy: { phosphor: "clipboard-text", weight: "regular" },
    cluster: { phosphor: null, weight: "bold", custom: { glyph: "cluster" } },
    clusterVariants: {
      outlined: { phosphor: null, weight: "bold", custom: { glyph: "cluster", variant: "outlined" } },
      filled: { phosphor: null, weight: "bold", custom: { glyph: "cluster", variant: "filled" } },
    },
    serviceAccount: { phosphor: "squares-four", weight: "bold" },
    audit: { phosphor: "list-dashes", weight: "bold" },
    department: { phosphor: "garage", weight: "bold" },
    report: { phosphor: "table", weight: "bold" },
    aiApplication: { phosphor: "app-window", weight: "bold" },
    user: { phosphor: "user", weight: "bold" },
    workload: { phosphor: null, weight: "bold", custom: { glyph: "workload" } },
    // New from catalog gaps:
    network: { phosphor: "network", weight: "bold" },
    users: { phosphor: "users", weight: "bold" },
    // Distributed-training "Master" step (left rail of training-distributed wizard).
    master: { phosphor: "graduation-cap", weight: "bold" },
    // Distributed-inference "Master" step. Phosphor `hard-hat` matched the fa-helmet-safety visual
    // UX expects; flagged previously as a "genuine gap" but Phosphor v2 ships it.
    inferenceMaster: { phosphor: "hard-hat", weight: "bold" },
    // Slack integration (Notifications). Phosphor ships `slack-logo` as a brand glyph.
    slack: { phosphor: "slack-logo", weight: "bold" },
  },
} as const satisfies Record<string, Record<string, IIconEntry | Record<string, IIconEntry>>>;

export const toQuasarIconName = (entry: IIconEntry): string => {
  if (entry.phosphor === null) {
    if (!entry.custom) return "";

    const variantPrefix = entry.custom.variant ? `${entry.custom.variant}-` : "";
    return `runai-${variantPrefix}${entry.custom.glyph}`;
  }
  return entry.weight === "regular" ? `ph ph-${entry.phosphor}` : `ph-${entry.weight} ph-${entry.phosphor}`;
};
