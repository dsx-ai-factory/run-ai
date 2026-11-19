/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
// Native number inputs permit exponent notation; these editors use plain numbers.
export function numericKeydown(event: KeyboardEvent, whole = false) {
  if (event.ctrlKey || event.metaKey || event.key.length !== 1) return;
  if (!(whole ? /^[0-9]$/ : /^[0-9.]$/).test(event.key)) event.preventDefault();
}
export function numericPaste(event: ClipboardEvent, whole = false) {
  const value = event.clipboardData?.getData("text");
  if (value == null) return;
  const valid = whole ? /^\d*$/.test(value) : /^[0-9.]*$/.test(value) && value.split(".").length <= 2;
  if (!valid) event.preventDefault();
}
export function numericBeforeInput(event: InputEvent, whole = false) {
  if (event.data && !(whole ? /^\d+$/ : /^[0-9.]+$/).test(event.data)) event.preventDefault();
}
