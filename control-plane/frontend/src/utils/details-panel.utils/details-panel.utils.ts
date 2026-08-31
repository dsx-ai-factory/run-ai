/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IDetailsPanelItem } from "@/models/details-panel.model";

function includesCI(value: string, term: string): boolean {
  return value.toLowerCase().includes(term.toLowerCase());
}

/** True when any string leaf inside `obj` contains `term` (case-insensitive). */
export function objectMatchesTerm(obj: unknown, term: string): boolean {
  if (!term) return true;
  if (obj === null || obj === undefined) return false;
  if (typeof obj === "string") return includesCI(obj, term);
  if (typeof obj === "number" || typeof obj === "boolean") return includesCI(String(obj), term);
  if (Array.isArray(obj)) return obj.some((v) => objectMatchesTerm(v, term));
  if (typeof obj === "object") return Object.values(obj).some((v) => objectMatchesTerm(v, term));
  return false;
}

/**
 * Filter a single panel item against a search term.
 * - If `value` is a string array, only elements containing the term are kept (the row is
 *   trimmed to the matching elements). If the label or link text matches, the whole row is
 *   kept untouched.
 * - If `value` is a string, the row is kept when label, link text, or value contains the term.
 * Returns `null` when nothing matches.
 */
export function filterItem(item: IDetailsPanelItem, term: string): IDetailsPanelItem | null {
  if (!term) return item;
  const labelOrLinkMatches = includesCI(item.label, term) || (!!item.link?.text && includesCI(item.link.text, term));

  if (Array.isArray(item.value)) {
    if (labelOrLinkMatches) return item;
    const matchingValues = item.value.filter((v) => includesCI(v, term));
    return matchingValues.length ? { ...item, value: matchingValues } : null;
  }

  if (labelOrLinkMatches || includesCI(item.value, term)) return item;
  return null;
}

/** Filter a list of panel items, applying per-element array filtering inside each item. */
export function filterItems(items: IDetailsPanelItem[], term: string): IDetailsPanelItem[] {
  if (!term) return items;
  return items.map((it) => filterItem(it, term)).filter((it): it is IDetailsPanelItem => it !== null);
}

/** Filter a flat array of strings, keeping only the entries that contain the term. */
export function filterStrings(strings: string[], term: string): string[] {
  if (!term) return strings;
  return strings.filter((s) => includesCI(s, term));
}

/** Filter an array of objects — keep an object whole when any leaf string contains the term. */
export function filterObjects<T>(objects: T[], term: string): T[] {
  if (!term) return objects;
  return objects.filter((o) => objectMatchesTerm(o, term));
}

/** True when an item matches the term (used by section-level `customMatch`). */
export function matchesItem(item: IDetailsPanelItem, term: string): boolean {
  return filterItem(item, term) !== null;
}

/** @deprecated use `filterItems` */
export function filterItemsByTerm(items: IDetailsPanelItem[], term: string): IDetailsPanelItem[] {
  return filterItems(items, term);
}
