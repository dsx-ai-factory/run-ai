/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { describe, expect, it, vi } from "vitest";
import { numericKeydown, numericPaste, numericBeforeInput } from "./scheduling-numeric-input";
describe("scheduling numeric entry", () => {
  it.each(["abc", "2e1", "-1", "+2", "12abc", "1.2.3"])(
    "rejects invalid pasted text %s without rewriting it",
    (value) => {
      const preventDefault = vi.fn();
      numericPaste({ clipboardData: { getData: () => value }, preventDefault } as unknown as ClipboardEvent);
      expect(preventDefault).toHaveBeenCalledOnce();
    },
  );
  it.each(["0", "20", "", "0.5", ".5"])("allows plain decimal text %s", (value) => {
    const preventDefault = vi.fn();
    numericPaste({ clipboardData: { getData: () => value }, preventDefault } as unknown as ClipboardEvent);
    expect(preventDefault).not.toHaveBeenCalled();
  });
  it("rejects fractional weight paste but preserves keyboard shortcuts", () => {
    const preventDefault = vi.fn();
    numericPaste({ clipboardData: { getData: () => "1.5" }, preventDefault } as unknown as ClipboardEvent, true);
    expect(preventDefault).toHaveBeenCalledOnce();
    const shortcut = new KeyboardEvent("keydown", { key: "a", ctrlKey: true, cancelable: true });
    numericKeydown(shortcut, true);
    expect(shortcut.defaultPrevented).toBe(false);
    const insertion = new InputEvent("beforeinput", { data: "abc", cancelable: true });
    numericBeforeInput(insertion);
    expect(insertion.defaultPrevented).toBe(true);
  });
});
