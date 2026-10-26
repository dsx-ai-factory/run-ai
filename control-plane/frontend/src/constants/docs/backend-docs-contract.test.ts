/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { readFileSync, readdirSync } from "node:fs";
import { join, relative, resolve } from "node:path";
import { describe, expect, it } from "vitest";
import { DOCS_CATALOG_V1, DOCS_CATALOG_V2 } from "./catalogs";

const controlPlaneDirectory = resolve("..");
const excludedDirectories = new Set(["frontend", "node_modules", "vendor", "generated", "_build", "bin"]);

function findBackendDocsReferences(directory: string): { file: string; key: string }[] {
  // eslint-disable-next-line security/detect-non-literal-fs-filename -- Only traverses repository directories; symlinks are not followed.
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const file = join(directory, entry.name);
    if (entry.isDirectory()) {
      return entry.name.startsWith(".") || excludedDirectories.has(entry.name) ? [] : findBackendDocsReferences(file);
    }
    if (!entry.isFile() || !entry.name.endsWith(".go") || entry.name.endsWith("_test.go")) return [];

    // eslint-disable-next-line security/detect-non-literal-fs-filename -- File paths come from the repository traversal above.
    const source = readFileSync(file, "utf8").replace(/\\"/g, '"');
    return Array.from(source.matchAll(/data-doc-key\s*=\s*["']([^"']*)["']/g), ([, key]) => ({
      file: relative(controlPlaneDirectory, file),
      key,
    }));
  });
}

describe("backend documentation references", () => {
  const references = findBackendDocsReferences(controlPlaneDirectory);

  it.each([
    ["v1", DOCS_CATALOG_V1],
    ["v2", DOCS_CATALOG_V2],
  ] as const)("resolves every backend description key to an available %s destination", (_, catalog) => {
    expect(references.length).toBeGreaterThan(0);
    const paths = new Map<string, string | null>(Object.entries(catalog));
    const unresolved = references.filter(({ key }) => !paths.get(key));

    expect(unresolved).toEqual([]);
  });
});
