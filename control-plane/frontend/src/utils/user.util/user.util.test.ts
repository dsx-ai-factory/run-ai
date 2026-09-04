/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import type { IPaginationFilter } from "@/models/filter.model";
import { userUtils } from "@/utils/user.util/user.util";

describe("userUtils", () => {
  describe("getFilterByArrays", () => {
    it("should include isLocal==true when isLocalUsers is true", () => {
      const filters: IPaginationFilter = {
        filterBy: ["type!=admin", "lastUpdated>=2023-01-01"],
      };
      const isLocalUsers = true;

      const result = userUtils.getFilterByArrays(filters, isLocalUsers);

      expect(result).toContain("isLocal==true");
      expect(result).toHaveLength(2); // type!=admin and isLocal==true
    });

    it("should include isLocal==false when type!= is present", () => {
      const filters: IPaginationFilter = {
        filterBy: ["type!=true", "lastUpdated>=2023-01-01"],
      };
      const isLocalUsers = false;

      const result = userUtils.getFilterByArrays(filters, isLocalUsers);

      expect(result).toContain("isLocal==false");
      expect(result).toHaveLength(2); // type!=admin and isLocal==false
    });

    it("should not include isLocal when type!= is not present", () => {
      const filters: IPaginationFilter = {
        filterBy: ["lastUpdated>=2023-01-01"],
      };
      const isLocalUsers = false;

      const result = userUtils.getFilterByArrays(filters, isLocalUsers);

      expect(result).not.toContain("isLocal");
      expect(result).toHaveLength(1); // Only lastUpdated filter remains
    });

    it("should handle multiple date filters and convert them to equality filters", () => {
      const filters: IPaginationFilter = {
        filterBy: [
          "lastUpdated>=2023-01-01",
          "lastUpdated<=2023-01-01",
          "creationTime<=2023-12-31",
          "creationTime>=2023-12-31",
          "type!=admin",
        ],
      };
      const isLocalUsers = false;

      const result = userUtils.getFilterByArrays(filters, isLocalUsers);

      // Check that the filters are transformed correctly
      expect(result).toContain("lastUpdated==2023-01-01");
      expect(result).toContain("creationTime==2023-12-31");
      expect(result).toHaveLength(3); // two equality filters and one type!= filter
    });
  });

  describe("getUserDisplayLabel", () => {
    it("returns email-first label followed by full name when both names are present", () => {
      expect(
        userUtils.getUserDisplayLabel({ username: "alice@run.ai", firstName: "Alice", lastName: "Cooper" }),
      ).toEqual({
        label: "alice@run.ai Alice Cooper",
        primaryLabel: "alice@run.ai",
        secondaryLabel: "Alice Cooper",
      });
    });

    it("uses firstName only when lastName is missing", () => {
      expect(userUtils.getUserDisplayLabel({ username: "alice@run.ai", firstName: "Alice" })).toEqual({
        label: "alice@run.ai Alice",
        primaryLabel: "alice@run.ai",
        secondaryLabel: "Alice",
      });
    });

    it("uses lastName only when firstName is missing", () => {
      expect(userUtils.getUserDisplayLabel({ username: "alice@run.ai", lastName: "Cooper" })).toEqual({
        label: "alice@run.ai Cooper",
        primaryLabel: "alice@run.ai",
        secondaryLabel: "Cooper",
      });
    });

    it("falls back to username only when both names are missing", () => {
      expect(userUtils.getUserDisplayLabel({ username: "alice@run.ai" })).toEqual({
        label: "alice@run.ai",
        primaryLabel: "alice@run.ai",
      });
    });

    it("falls back to username only when names are null", () => {
      expect(userUtils.getUserDisplayLabel({ username: "alice@run.ai", firstName: null, lastName: null })).toEqual({
        label: "alice@run.ai",
        primaryLabel: "alice@run.ai",
      });
    });

    it("falls back to username only when names are empty strings", () => {
      expect(userUtils.getUserDisplayLabel({ username: "alice@run.ai", firstName: "", lastName: "" })).toEqual({
        label: "alice@run.ai",
        primaryLabel: "alice@run.ai",
      });
    });

    it("ignores empty firstName and uses lastName only", () => {
      expect(userUtils.getUserDisplayLabel({ username: "alice@run.ai", firstName: "", lastName: "Cooper" })).toEqual({
        label: "alice@run.ai Cooper",
        primaryLabel: "alice@run.ai",
        secondaryLabel: "Cooper",
      });
    });
  });

  describe("getFilterByWithEqualsDates", () => {
    it("should convert multiple date filters to equality filters", () => {
      const filters: IPaginationFilter = {
        filterBy: [
          "lastUpdated>=2023-01-01",
          "lastUpdated<=2023-01-01",
          "creationTime<=2023-12-31",
          "creationTime>=2023-12-31",
          "lastLogin>=2023-06-01",
          "lastLogin<=2023-06-01",
        ],
      };

      const result = userUtils.getFilterByWithEqualsDates(filters);

      expect(result).toContain("lastUpdated==2023-01-01");
      expect(result).toContain("creationTime==2023-12-31");
      expect(result).toContain("lastLogin==2023-06-01");
      expect(result).not.toContain("lastUpdated>=2023-01-01");
      expect(result).not.toContain("creationTime<=2023-12-31");
      expect(result).not.toContain("lastLogin>=2023-06-01");
    });

    it("should return the same filters if no date filters are present", () => {
      const filters: IPaginationFilter = {
        filterBy: ["type!=admin", "status==active"],
      };

      const result = userUtils.getFilterByWithEqualsDates(filters);

      expect(result).toEqual(filters.filterBy); // should remain unchanged
    });

    it("should not modify filters with only one date instance", () => {
      const filters: IPaginationFilter = {
        filterBy: ["lastUpdated>=2023-01-01"],
      };

      const result = userUtils.getFilterByWithEqualsDates(filters);

      expect(result).toEqual(filters.filterBy); // should remain unchanged
    });
  });
});
