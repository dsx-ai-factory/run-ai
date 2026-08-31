/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it } from "vitest";
import { EResourceType } from "@/models/resource.model";
import { resourceUtil } from "@/utils/resource.util";
import type { Resources } from "@/swagger-models/org-unit-service-client";

describe("resource.util", () => {
  describe("sortResourcesByNodePoolOrder", () => {
    const makeResource = (name: string) => ({ nodePool: { id: name, name } }) as Resources;

    const poolA = makeResource("pool-a");
    const poolB = makeResource("pool-b");
    const poolC = makeResource("pool-c");
    const poolD = makeResource("pool-d");

    it("should sort resources by the given priority order", () => {
      const result = resourceUtil.sortResourcesByNodePoolOrder([poolC, poolA, poolB], ["pool-a", "pool-b", "pool-c"]);
      expect(result).toEqual([poolA, poolB, poolC]);
    });

    it("should place priority items before non-priority items", () => {
      const result = resourceUtil.sortResourcesByNodePoolOrder([poolD, poolA, poolC], ["pool-a", "pool-c"]);
      expect(result).toEqual([poolA, poolC, poolD]);
    });

    it("should preserve relative order of non-priority items", () => {
      const result = resourceUtil.sortResourcesByNodePoolOrder([poolD, poolC, poolA], ["pool-a"]);
      expect(result).toEqual([poolA, poolD, poolC]);
    });

    it("should return the same order when no items match priority", () => {
      const result = resourceUtil.sortResourcesByNodePoolOrder([poolC, poolB, poolA], ["pool-x"]);
      expect(result).toEqual([poolC, poolB, poolA]);
    });

    it("should return empty array when given empty resources", () => {
      const result = resourceUtil.sortResourcesByNodePoolOrder([], ["pool-a"]);
      expect(result).toEqual([]);
    });

    it("should not mutate the original array", () => {
      const original = [poolC, poolA, poolB];
      resourceUtil.sortResourcesByNodePoolOrder(original, ["pool-a", "pool-b", "pool-c"]);
      expect(original).toEqual([poolC, poolA, poolB]);
    });
  });

  describe("getResourceDisplayValue", () => {
    it("should return 'Unlimited' if value is null and the resource is CPU", () => {
      expect(resourceUtil.getResourceDisplayValue(null, EResourceType.CPU)).toEqual("Unlimited");
    });

    it("should return 'Unlimited' if value is null and the resource is MEMORY", () => {
      expect(resourceUtil.getResourceDisplayValue(null, EResourceType.MEMORY)).toEqual("Unlimited");
    });

    it("should not return 'Unlimited' if the resource is GPU", () => {
      expect(resourceUtil.getResourceDisplayValue(null, EResourceType.GPU)).not.toEqual("Unlimited");
    });

    it("should return '0.00' if value is null and the resource is GPU", () => {
      expect(resourceUtil.getResourceDisplayValue(null, EResourceType.GPU)).toEqual("0.00");
    });

    it("should return 'Unlimited' if value is -1 and the resource is GPU", () => {
      expect(resourceUtil.getResourceDisplayValue(-1, EResourceType.GPU)).toEqual("Unlimited");
    });

    it("should return 'Unlimited' if value is -1 and the resource is GPU_LIMIT", () => {
      expect(resourceUtil.getResourceDisplayValue(-1, EResourceType.GPU_LIMIT)).toEqual("Unlimited");
    });

    it("should return the value.toFixed(2)", () => {
      expect(resourceUtil.getResourceDisplayValue(5.5555555, EResourceType.GPU)).toEqual("5.56");
    });
  });
});
