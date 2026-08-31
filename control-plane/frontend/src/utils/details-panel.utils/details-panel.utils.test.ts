/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import type { IDetailsPanelItem } from "@/models/details-panel.model";
import {
  filterItem,
  filterItems,
  filterStrings,
  filterObjects,
  objectMatchesTerm,
  matchesItem,
  filterItemsByTerm,
} from "@/utils/details-panel.utils";

function item(overrides: Partial<IDetailsPanelItem> = {}): IDetailsPanelItem {
  return { label: "Label", value: "value", ...overrides };
}

describe("filterItem", () => {
  it("returns the item unchanged when the term is empty", () => {
    const it = item({ label: "Project", value: ["a", "b"] });
    expect(filterItem(it, "")).toBe(it);
  });

  it("returns the item when the label matches (string value)", () => {
    const it = item({ label: "Project", value: "p100" });
    expect(filterItem(it, "proj")).toBe(it);
  });

  it("returns the item when the value matches (string value)", () => {
    const it = item({ label: "Project", value: "p100" });
    expect(filterItem(it, "100")).toBe(it);
  });

  it("matches case-insensitively", () => {
    expect(filterItem(item({ label: "Project", value: "P100" }), "p100")).not.toBeNull();
    expect(filterItem(item({ label: "PROJECT", value: "p100" }), "project")).not.toBeNull();
  });

  it("returns null when nothing matches", () => {
    expect(filterItem(item({ label: "Project", value: "p100" }), "zzz")).toBeNull();
  });

  it("keeps the row but trims an array value to only matching entries", () => {
    const it = item({ label: "Tools", value: ["Jupyter", "VSCode", "TensorBoard"] });
    const result = filterItem(it, "code");
    expect(result).not.toBeNull();
    expect(result?.value).toEqual(["VSCode"]);
  });

  it("returns null when no array entry matches AND label/link don't match", () => {
    const it = item({ label: "Tools", value: ["Jupyter", "VSCode"] });
    expect(filterItem(it, "xyz")).toBeNull();
  });

  it("keeps the whole array when the label matches", () => {
    const it = item({ label: "Tools", value: ["Jupyter", "VSCode"] });
    const result = filterItem(it, "tool");
    expect(result).toBe(it);
    expect(result?.value).toEqual(["Jupyter", "VSCode"]);
  });

  it("matches against link text", () => {
    const it = item({ label: "Project", value: "—", link: { text: "p100", href: "/projects" } });
    expect(filterItem(it, "p100")).toBe(it);
  });

  it("keeps the whole array when link text matches", () => {
    const it = item({
      label: "Tools",
      value: ["Jupyter", "VSCode"],
      link: { text: "open tools", href: "/x" },
    });
    const result = filterItem(it, "open");
    expect(result).toBe(it);
    expect(result?.value).toEqual(["Jupyter", "VSCode"]);
  });
});

describe("filterItems", () => {
  it("returns the input array reference when the term is empty", () => {
    const items = [item({ label: "A" }), item({ label: "B" })];
    expect(filterItems(items, "")).toBe(items);
  });

  it("drops items with no match", () => {
    const items = [item({ label: "Project", value: "p100" }), item({ label: "Cluster", value: "us-east" })];
    expect(filterItems(items, "p100")).toEqual([items[0]]);
  });

  it("trims array values within kept items", () => {
    const items = [item({ label: "Tools", value: ["Jupyter", "VSCode", "TensorBoard"] })];
    const result = filterItems(items, "board");
    expect(result).toHaveLength(1);
    expect(result[0].value).toEqual(["TensorBoard"]);
  });

  it("returns an empty array when no item matches", () => {
    const items = [item({ label: "A", value: "x" }), item({ label: "B", value: "y" })];
    expect(filterItems(items, "zzz")).toEqual([]);
  });
});

describe("filterStrings", () => {
  it("returns the input array reference when the term is empty", () => {
    const strings = ["a", "b"];
    expect(filterStrings(strings, "")).toBe(strings);
  });

  it("keeps only entries containing the term", () => {
    expect(filterStrings(["FOO=1", "BAR=2", "Foobar=3"], "foo")).toEqual(["FOO=1", "Foobar=3"]);
  });

  it("returns an empty array when nothing matches", () => {
    expect(filterStrings(["a", "b"], "zzz")).toEqual([]);
  });
});

describe("objectMatchesTerm", () => {
  it("returns true when the term is empty", () => {
    expect(objectMatchesTerm({ name: "anything" }, "")).toBe(true);
  });

  it("returns false for null / undefined", () => {
    expect(objectMatchesTerm(null, "foo")).toBe(false);
    expect(objectMatchesTerm(undefined, "foo")).toBe(false);
  });

  it("matches primitive strings, numbers, booleans", () => {
    expect(objectMatchesTerm("Jupyter", "jupy")).toBe(true);
    expect(objectMatchesTerm(8080, "808")).toBe(true);
    expect(objectMatchesTerm(true, "true")).toBe(true);
    expect(objectMatchesTerm(false, "zzz")).toBe(false);
  });

  it("walks nested objects and arrays", () => {
    const obj = { name: "url-master", details: { ports: [8080, 8888], tool: "VSCode" } };
    expect(objectMatchesTerm(obj, "code")).toBe(true);
    expect(objectMatchesTerm(obj, "8888")).toBe(true);
    expect(objectMatchesTerm(obj, "master")).toBe(true);
    expect(objectMatchesTerm(obj, "absent")).toBe(false);
  });

  it("treats symbols / functions as non-matches", () => {
    expect(objectMatchesTerm(Symbol("x"), "x")).toBe(false);
    expect(objectMatchesTerm(() => "x", "x")).toBe(false);
  });
});

describe("filterObjects", () => {
  it("returns the input array reference when the term is empty", () => {
    const xs = [{ name: "a" }, { name: "b" }];
    expect(filterObjects(xs, "")).toBe(xs);
  });

  it("keeps an object whole when any leaf string contains the term", () => {
    const xs = [
      { name: "conn-a", port: 80, tool: "Jupyter" },
      { name: "conn-b", port: 81, tool: "VSCode" },
    ];
    expect(filterObjects(xs, "code")).toEqual([xs[1]]);
  });

  it("returns an empty array when no object matches", () => {
    const xs = [{ a: 1 }, { b: 2 }];
    expect(filterObjects(xs, "zzz")).toEqual([]);
  });
});

describe("matchesItem", () => {
  it("returns true when something matches", () => {
    expect(matchesItem(item({ label: "Project", value: "p100" }), "100")).toBe(true);
  });

  it("returns false when nothing matches", () => {
    expect(matchesItem(item({ label: "Project", value: "p100" }), "zzz")).toBe(false);
  });
});

describe("filterItemsByTerm (deprecated alias)", () => {
  it("delegates to filterItems", () => {
    const items = [item({ label: "Project", value: "p100" }), item({ label: "Cluster", value: "us-east" })];
    expect(filterItemsByTerm(items, "p100")).toEqual([items[0]]);
  });
});
