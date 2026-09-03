/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export const storageUtil = {
  save,
  get,
  remove,
};

function get<T>(key: string): T | null {
  try {
    const data: string | null = localStorage.getItem(key);
    return data && data !== "undefined" ? JSON.parse(data) : null;
  } catch (error) {
    console.error(`Error retrieving item with key ${key}:`, error);
    return null;
  }
}

function save<T>(key: string, item: T): void {
  localStorage.setItem(key, JSON.stringify(item));
}

function remove(key: string): void {
  localStorage.removeItem(key);
}
