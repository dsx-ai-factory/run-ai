/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export default {
  name: "phosphor",
  type: {
    positive: "ph-bold ph-check-circle",
    negative: "ph-bold ph-warning-circle",
    info: "ph-bold ph-info",
    warning: "ph-bold ph-warning",
  },
  arrow: {
    up: "ph-bold ph-arrow-up",
    right: "ph-bold ph-arrow-right",
    down: "ph-bold ph-arrow-down",
    left: "ph-bold ph-arrow-left",
    dropdown: "ph-fill ph-caret-down",
  },
  chevron: {
    left: "ph-bold ph-caret-left",
    right: "ph-bold ph-caret-right",
  },
  colorPicker: {
    spectrum: "ph-bold ph-eyedropper",
    tune: "ph-bold ph-sliders",
    palette: "ph-bold ph-swatches",
  },
  pullToRefresh: {
    icon: "ph-bold ph-arrows-clockwise",
  },
  carousel: {
    left: "ph-bold ph-caret-left",
    right: "ph-bold ph-caret-right",
    up: "ph-bold ph-caret-up",
    down: "ph-bold ph-caret-down",
    navigationIcon: "ph-bold ph-circle",
  },
  chip: {
    remove: "ph-fill ph-x-circle",
    selected: "ph-bold ph-check",
  },
  datetime: {
    arrowLeft: "ph-bold ph-caret-left",
    arrowRight: "ph-bold ph-caret-right",
    now: "ph-bold ph-clock",
    today: "ph-bold ph-calendar-check",
  },
  editor: {
    bold: "ph-bold ph-text-bolder",
    italic: "ph-bold ph-text-italic",
    strikethrough: "ph-bold ph-text-strikethrough",
    underline: "ph-bold ph-text-underline",
    unorderedList: "ph-bold ph-list-bullets",
    orderedList: "ph-bold ph-list-numbers",
    subscript: "ph-bold ph-text-subscript",
    superscript: "ph-bold ph-text-superscript",
    hyperlink: "ph-bold ph-link",
    toggleFullscreen: "ph-bold ph-arrows-out",
    quote: "ph-bold ph-quotes",
    left: "ph-bold ph-text-align-left",
    center: "ph-bold ph-text-align-center",
    right: "ph-bold ph-text-align-right",
    justify: "ph-bold ph-text-align-justify",
    print: "ph-bold ph-printer",
    outdent: "ph-bold ph-text-outdent",
    indent: "ph-bold ph-text-indent",
    removeFormat: "ph-bold ph-eraser",
    formatting: "ph-bold ph-text-h",
    fontSize: "ph-bold ph-text-aa",
    align: "ph-bold ph-text-align-left",
    hr: "ph-bold ph-minus-square",
    undo: "ph-bold ph-arrow-counter-clockwise",
    redo: "ph-bold ph-arrow-clockwise",
    heading: "ph-bold ph-text-h",
    code: "ph-bold ph-code",
    size: "ph-bold ph-text-aa",
    font: "ph-bold ph-text-t",
    viewSource: "ph-bold ph-code",
  },
  expansionItem: {
    icon: "ph-bold ph-caret-down",
    denseIcon: "ph-bold ph-caret-down",
  },
  fab: {
    icon: "ph-bold ph-plus",
    activeIcon: "ph-bold ph-x",
  },
  field: {
    clear: "ph-bold ph-x-circle",
    error: "ph-bold ph-warning-circle",
  },
  pagination: {
    first: "ph-bold ph-skip-back",
    prev: "ph-bold ph-caret-left",
    next: "ph-bold ph-caret-right",
    last: "ph-bold ph-skip-forward",
  },
  rating: {
    icon: "ph-bold ph-star",
  },
  stepper: {
    done: "ph-bold ph-check",
    active: "ph-fill ph-pencil-simple",
    error: "ph-fill ph-warning",
  },
  tabs: {
    left: "ph-bold ph-caret-left",
    right: "ph-bold ph-caret-right",
    up: "ph-bold ph-caret-up",
    down: "ph-bold ph-caret-down",
  },
  table: {
    arrowUp: "ph-bold ph-arrow-up",
    warning: "ph-bold ph-warning",
    firstPage: "ph-bold ph-caret-double-left",
    prevPage: "ph-bold ph-caret-left",
    nextPage: "ph-bold ph-caret-right",
    lastPage: "ph-bold ph-caret-double-right",
  },
  tree: {
    // q-tree rotates this single icon between collapsed (right-pointing) and expanded (down-pointing).
    // Filled caret matches the Figma scope-tree spec — light weight rendered as outlined triangles.
    icon: "ph-fill ph-caret-right",
  },
  uploader: {
    done: "ph-bold ph-check",
    clear: "ph-bold ph-x",
    add: "ph-bold ph-plus-square",
    upload: "ph-bold ph-upload-simple",
    removeQueue: "ph-bold ph-list-dashes",
    removeUploaded: "ph-bold ph-clipboard-text",
  },
};
