/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it } from "vitest";
import { AssetKind } from "@/swagger-models/assets-service-client";
import type { SupersetRulesAllOfStorage } from "@/swagger-models/policy-service-client";
import { dataSourceUtil } from "./data-source.util";

describe("data-source.util", () => {
  describe("filterDataSourcesByPolicyRules", () => {
    it("should return the original set when datasourcesRules is undefined", () => {
      const inputSet = new Set([AssetKind.Pvc, AssetKind.Nfs]);
      const result = dataSourceUtil.filterDataSourcesByPolicyRules(inputSet, undefined);
      expect(result).toEqual(inputSet);
    });

    it("should return the original set when datasourcesRules is null", () => {
      const inputSet = new Set([AssetKind.Pvc, AssetKind.Nfs]);
      const result = dataSourceUtil.filterDataSourcesByPolicyRules(inputSet, null);
      expect(result).toEqual(inputSet);
    });

    it("should filter out data sources based on policy rules", () => {
      const inputSet = new Set([AssetKind.Pvc, AssetKind.Nfs, AssetKind.S3]);
      const rules: SupersetRulesAllOfStorage = {
        pvc: { instances: { canAdd: false } },
        nfs: { instances: { canAdd: true } },
        s3: { instances: { canAdd: false } },
      };

      const result = dataSourceUtil.filterDataSourcesByPolicyRules(inputSet, rules);
      expect(result).toEqual(new Set([AssetKind.Nfs]));
    });

    it("should handle empty input set", () => {
      const inputSet = new Set<AssetKind>();
      const rules: SupersetRulesAllOfStorage = {
        pvc: { instances: { canAdd: false } },
      };

      const result = dataSourceUtil.filterDataSourcesByPolicyRules(inputSet, rules);
      expect(result).toEqual(new Set());
    });

    it("should handle rules with missing instances property", () => {
      const inputSet = new Set([AssetKind.Pvc, AssetKind.Nfs]);
      const rules = {
        pvc: {},
        nfs: { instances: { canAdd: false } },
      } as SupersetRulesAllOfStorage;

      const result = dataSourceUtil.filterDataSourcesByPolicyRules(inputSet, rules);
      expect(result).toEqual(new Set([AssetKind.Pvc]));
    });

    it("should handle rules with undefined canAdd", () => {
      const inputSet = new Set([AssetKind.Pvc, AssetKind.Nfs]);
      const rules = {
        pvc: { instances: {} },
        nfs: { instances: { canAdd: false } },
      } as SupersetRulesAllOfStorage;

      const result = dataSourceUtil.filterDataSourcesByPolicyRules(inputSet, rules);
      expect(result).toEqual(new Set([AssetKind.Pvc]));
    });

    it("should preserve the type when using generics", () => {
      type ValidDataSources = typeof AssetKind.Pvc | typeof AssetKind.Nfs;
      const inputSet = new Set<ValidDataSources>([AssetKind.Pvc, AssetKind.Nfs]);
      const rules: SupersetRulesAllOfStorage = {
        pvc: { instances: { canAdd: false } },
      };

      const result = dataSourceUtil.filterDataSourcesByPolicyRules<ValidDataSources>(inputSet, rules);
      expect(result).toEqual(new Set([AssetKind.Nfs]));
    });

    it("should handle all supported asset kinds", () => {
      const inputSet = new Set([
        AssetKind.Pvc,
        AssetKind.Nfs,
        AssetKind.S3,
        AssetKind.Git,
        AssetKind.HostPath,
        AssetKind.SecretVolume,
        AssetKind.ConfigMap,
      ]);
      const rules: SupersetRulesAllOfStorage = {
        pvc: { instances: { canAdd: false } },
        nfs: { instances: { canAdd: true } },
        s3: { instances: { canAdd: false } },
        git: { instances: { canAdd: true } },
        hostPath: { instances: { canAdd: false } },
        secretVolume: { instances: { canAdd: true } },
        configMapVolume: { instances: { canAdd: false } },
      };

      const result = dataSourceUtil.filterDataSourcesByPolicyRules(inputSet, rules);
      expect(result).toEqual(new Set([AssetKind.Nfs, AssetKind.Git, AssetKind.SecretVolume]));
    });
  });
});
