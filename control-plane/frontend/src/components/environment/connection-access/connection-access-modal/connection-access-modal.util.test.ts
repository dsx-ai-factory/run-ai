/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { convertNamesToItems } from "./connection-access-modal.util";

import { EToolLabel } from "./connection-access-modal.model";
import { describe, it, expect } from "vitest";

describe("connection-access-modal.util", () => {
  describe("convertNamesToItems", () => {
    it("should convert names to items correctly", () => {
      const names = ["Alice", "Bob"];
      const result = convertNamesToItems(names, EToolLabel.GroupName);

      expect(result).toEqual([
        { text: "Alice", label: EToolLabel.GroupName, removable: true },
        { text: "Bob", label: EToolLabel.GroupName, removable: true },
      ]);
    });

    it("should set removable to false when there is only one item in the array", () => {
      const names = ["Alice"];
      const result = convertNamesToItems(names, EToolLabel.GroupName);

      expect(result).toEqual([{ text: "Alice", label: EToolLabel.GroupName, removable: false }]);
    });
  });
});
