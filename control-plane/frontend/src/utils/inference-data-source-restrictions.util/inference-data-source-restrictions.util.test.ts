/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, beforeEach, vi } from "vitest";
import { inferenceDataSourceRestrictionsUtil } from "./inference-data-source-restrictions.util";
import type { IDrawerListItem } from "@/models/workload-creation-drawer-model";
import { AssetKind } from "@/swagger-models/assets-service-client";
import { useClusterStore } from "@/stores/cluster.store";

// Mock the cluster store module
vi.mock("@/stores/cluster.store", () => ({
  useClusterStore: vi.fn(),
}));

describe("inference-data-source-restrictions.util", () => {
  const clusterId = "test-cluster-id";
  const mockClusterStore = {
    isClusterVersionSufficient: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    // Make useClusterStore return our mock
    vi.mocked(useClusterStore).mockReturnValue(mockClusterStore as unknown as ReturnType<typeof useClusterStore>);
  });

  describe("applyInferenceDataSourceRestrictions", () => {
    describe("when item has no kind", () => {
      it("should return item unchanged", () => {
        const items: IDrawerListItem[] = [
          {
            id: "1",
            name: "test-item",
            details: [],
          },
        ];

        const result = inferenceDataSourceRestrictionsUtil.applyInferenceDataSourceRestrictions(items, clusterId);

        expect(result).toEqual(items);
      });
    });

    describe("supported AssetKinds", () => {
      it("should allow PVC without restrictions", () => {
        const items: IDrawerListItem[] = [
          {
            id: "1",
            name: "test-pvc",
            kind: AssetKind.Pvc,
            details: [],
          },
        ];

        const result = inferenceDataSourceRestrictionsUtil.applyInferenceDataSourceRestrictions(items, clusterId);

        expect(result[0].disabled).toBeUndefined();
        expect(result[0].disabledInfo).toBeUndefined();
      });

      it("should allow ConfigMap in data source context", () => {
        const items: IDrawerListItem[] = [
          {
            id: "1",
            name: "test-configmap",
            kind: AssetKind.ConfigMap,
            details: [],
          },
        ];

        const result = inferenceDataSourceRestrictionsUtil.applyInferenceDataSourceRestrictions(items, clusterId, false);

        expect(result[0].disabled).toBeUndefined();
        expect(result[0].disabledInfo).toBeUndefined();
      });

      it("should allow Git in data source context", () => {
        const items: IDrawerListItem[] = [
          {
            id: "1",
            name: "test-git",
            kind: AssetKind.Git,
            details: [],
          },
        ];

        const result = inferenceDataSourceRestrictionsUtil.applyInferenceDataSourceRestrictions(items, clusterId, false);

        expect(result[0].disabled).toBeUndefined();
        expect(result[0].disabledInfo).toBeUndefined();
      });
    });

    describe("model store context restrictions", () => {
      it("should disable Git in model store context", () => {
        const items: IDrawerListItem[] = [
          {
            id: "1",
            name: "test-git",
            kind: AssetKind.Git,
            details: [],
          },
        ];

        const result = inferenceDataSourceRestrictionsUtil.applyInferenceDataSourceRestrictions(items, clusterId, true);

        expect(result[0].disabled).toBe(true);
        expect(result[0].disabledInfo).toBe(
          "This setup cannot be selected because the Inference model store does not support data sources of type Git",
        );
      });

      it("should disable ConfigMap in model store context", () => {
        const items: IDrawerListItem[] = [
          {
            id: "1",
            name: "test-configmap",
            kind: AssetKind.ConfigMap,
            details: [],
          },
        ];

        const result = inferenceDataSourceRestrictionsUtil.applyInferenceDataSourceRestrictions(items, clusterId, true);

        expect(result[0].disabled).toBe(true);
        expect(result[0].disabledInfo).toBe(
          "This setup cannot be selected because the Inference model store does not support data sources of type ConfigMap",
        );
      });

      it("should not affect PVC in model store context", () => {
        const items: IDrawerListItem[] = [
          {
            id: "1",
            name: "test-pvc",
            kind: AssetKind.Pvc,
            details: [],
          },
        ];

        const result = inferenceDataSourceRestrictionsUtil.applyInferenceDataSourceRestrictions(items, clusterId, true);

        expect(result[0].disabled).toBeUndefined();
        expect(result[0].disabledInfo).toBeUndefined();
      });
    });

    describe("unsupported AssetKinds", () => {
      it("should disable S3 data sources", () => {
        const items: IDrawerListItem[] = [
          {
            id: "1",
            name: "test-s3",
            kind: AssetKind.S3,
            details: [],
          },
        ];

        const result = inferenceDataSourceRestrictionsUtil.applyInferenceDataSourceRestrictions(items, clusterId);

        expect(result[0].disabled).toBe(true);
        expect(result[0].disabledInfo).toBe(
          "This setup cannot be selected because Inference workload does not support data sources of type S3 Bucket.",
        );
      });

      it("should disable SecretVolume data sources", () => {
        const items: IDrawerListItem[] = [
          {
            id: "1",
            name: "test-secret",
            kind: AssetKind.SecretVolume,
            details: [],
          },
        ];

        const result = inferenceDataSourceRestrictionsUtil.applyInferenceDataSourceRestrictions(items, clusterId);

        expect(result[0].disabled).toBe(true);
        expect(result[0].disabledInfo).toBe(
          "This setup cannot be selected because Inference workload does not support data sources of type Secret.",
        );
      });
    });

    describe("cluster version requirements", () => {
      it("should allow NFS when cluster version is sufficient", () => {
        mockClusterStore.isClusterVersionSufficient.mockReturnValue(true);

        const items: IDrawerListItem[] = [
          {
            id: "1",
            name: "test-nfs",
            kind: AssetKind.Nfs,
            details: [],
          },
        ];

        const result = inferenceDataSourceRestrictionsUtil.applyInferenceDataSourceRestrictions(items, clusterId);

        expect(mockClusterStore.isClusterVersionSufficient).toHaveBeenCalledWith(clusterId, "2.19");
        expect(result[0].disabled).toBeUndefined();
        expect(result[0].disabledInfo).toBeUndefined();
      });

      it("should disable NFS when cluster version is insufficient", () => {
        mockClusterStore.isClusterVersionSufficient.mockReturnValue(false);

        const items: IDrawerListItem[] = [
          {
            id: "1",
            name: "test-nfs",
            kind: AssetKind.Nfs,
            details: [],
          },
        ];

        const result = inferenceDataSourceRestrictionsUtil.applyInferenceDataSourceRestrictions(items, clusterId);

        expect(result[0].disabled).toBe(true);
        expect(result[0].disabledInfo).toBe(
          "Inference workloads don't support data sources of type NFS for the current version of the cluster.",
        );
      });

      it("should allow HostPath when cluster version is sufficient", () => {
        mockClusterStore.isClusterVersionSufficient.mockReturnValue(true);

        const items: IDrawerListItem[] = [
          {
            id: "1",
            name: "test-hostpath",
            kind: AssetKind.HostPath,
            details: [],
          },
        ];

        const result = inferenceDataSourceRestrictionsUtil.applyInferenceDataSourceRestrictions(items, clusterId);

        expect(mockClusterStore.isClusterVersionSufficient).toHaveBeenCalledWith(clusterId, "2.19");
        expect(result[0].disabled).toBeUndefined();
        expect(result[0].disabledInfo).toBeUndefined();
      });

      it("should disable HostPath when cluster version is insufficient", () => {
        mockClusterStore.isClusterVersionSufficient.mockReturnValue(false);

        const items: IDrawerListItem[] = [
          {
            id: "1",
            name: "test-hostpath",
            kind: AssetKind.HostPath,
            details: [],
          },
        ];

        const result = inferenceDataSourceRestrictionsUtil.applyInferenceDataSourceRestrictions(items, clusterId);

        expect(result[0].disabled).toBe(true);
        expect(result[0].disabledInfo).toBe(
          "Inference workloads don't support data sources of type Host path for the current version of the cluster.",
        );
      });

      it("should allow DataVolume when cluster version is sufficient", () => {
        mockClusterStore.isClusterVersionSufficient.mockReturnValue(true);

        const items: IDrawerListItem[] = [
          {
            id: "1",
            name: "test-datavolume",
            kind: AssetKind.DataVolume,
            details: [],
          },
        ];

        const result = inferenceDataSourceRestrictionsUtil.applyInferenceDataSourceRestrictions(items, clusterId);

        expect(mockClusterStore.isClusterVersionSufficient).toHaveBeenCalledWith(clusterId, "2.19");
        expect(result[0].disabled).toBeUndefined();
        expect(result[0].disabledInfo).toBeUndefined();
      });

      it("should disable DataVolume when cluster version is insufficient", () => {
        mockClusterStore.isClusterVersionSufficient.mockReturnValue(false);

        const items: IDrawerListItem[] = [
          {
            id: "1",
            name: "test-datavolume",
            kind: AssetKind.DataVolume,
            details: [],
          },
        ];

        const result = inferenceDataSourceRestrictionsUtil.applyInferenceDataSourceRestrictions(items, clusterId);

        expect(result[0].disabled).toBe(true);
        expect(result[0].disabledInfo).toBe(
          "Inference workloads don't support data sources of type Data volume for the current version of the cluster.",
        );
      });
    });

    describe("multiple items with mixed conditions", () => {
      it("should apply restrictions correctly to multiple items", () => {
        mockClusterStore.isClusterVersionSufficient.mockReturnValue(false);

        const items: IDrawerListItem[] = [
          {
            id: "1",
            name: "test-pvc",
            kind: AssetKind.Pvc,
            details: [],
          },
          {
            id: "2",
            name: "test-s3",
            kind: AssetKind.S3,
            details: [],
          },
          {
            id: "3",
            name: "test-nfs",
            kind: AssetKind.Nfs,
            details: [],
          },
          {
            id: "4",
            name: "test-git",
            kind: AssetKind.Git,
            details: [],
          },
        ];

        const result = inferenceDataSourceRestrictionsUtil.applyInferenceDataSourceRestrictions(items, clusterId, false);

        // PVC should be allowed
        expect(result[0].disabled).toBeUndefined();

        // S3 should be disabled (not supported)
        expect(result[1].disabled).toBe(true);
        expect(result[1].disabledInfo).toContain("S3 Bucket");

        // NFS should be disabled (version requirement)
        expect(result[2].disabled).toBe(true);
        expect(result[2].disabledInfo).toContain("current version of the cluster");

        // Git should be allowed in data source context
        expect(result[3].disabled).toBeUndefined();
      });

      it("should apply model store restrictions with version checks", () => {
        mockClusterStore.isClusterVersionSufficient.mockReturnValue(true);

        const items: IDrawerListItem[] = [
          {
            id: "1",
            name: "test-pvc",
            kind: AssetKind.Pvc,
            details: [],
          },
          {
            id: "2",
            name: "test-git",
            kind: AssetKind.Git,
            details: [],
          },
          {
            id: "3",
            name: "test-nfs",
            kind: AssetKind.Nfs,
            details: [],
          },
          {
            id: "4",
            name: "test-configmap",
            kind: AssetKind.ConfigMap,
            details: [],
          },
        ];

        const result = inferenceDataSourceRestrictionsUtil.applyInferenceDataSourceRestrictions(items, clusterId, true);

        // PVC should be allowed
        expect(result[0].disabled).toBeUndefined();

        // Git should be disabled in model store context
        expect(result[1].disabled).toBe(true);
        expect(result[1].disabledInfo).toContain("Inference model store");

        // NFS should be allowed (version is sufficient)
        expect(result[2].disabled).toBeUndefined();

        // ConfigMap should be disabled in model store context
        expect(result[3].disabled).toBe(true);
        expect(result[3].disabledInfo).toContain("Inference model store");
      });
    });

    describe("edge cases", () => {
      it("should handle empty array", () => {
        const items: IDrawerListItem[] = [];

        const result = inferenceDataSourceRestrictionsUtil.applyInferenceDataSourceRestrictions(items, clusterId);

        expect(result).toEqual([]);
      });

      it("should preserve existing item properties", () => {
        mockClusterStore.isClusterVersionSufficient.mockReturnValue(true);

        const items: IDrawerListItem[] = [
          {
            id: "1",
            name: "test-pvc",
            kind: AssetKind.Pvc,
            description: "Test description",
            details: [{ content: ["test"] }],
            createdAt: "2024-01-01",
          },
        ];

        const result = inferenceDataSourceRestrictionsUtil.applyInferenceDataSourceRestrictions(items, clusterId);

        expect(result[0]).toEqual({
          id: "1",
          name: "test-pvc",
          kind: AssetKind.Pvc,
          description: "Test description",
          details: [{ content: ["test"] }],
          createdAt: "2024-01-01",
        });
      });

      it("should not override existing disabled state if item passes all checks", () => {
        mockClusterStore.isClusterVersionSufficient.mockReturnValue(true);

        const items: IDrawerListItem[] = [
          {
            id: "1",
            name: "test-pvc",
            kind: AssetKind.Pvc,
            details: [],
            disabled: true,
            disabledInfo: "Already disabled for other reasons",
          },
        ];

        const result = inferenceDataSourceRestrictionsUtil.applyInferenceDataSourceRestrictions(items, clusterId);

        // Since PVC passes all checks, it returns the item as-is with original disabled state
        expect(result[0].disabled).toBe(true);
        expect(result[0].disabledInfo).toBe("Already disabled for other reasons");
      });
    });

    describe("default parameter values", () => {
      it("should use false as default for isModelStoreContext", () => {
        const items: IDrawerListItem[] = [
          {
            id: "1",
            name: "test-git",
            kind: AssetKind.Git,
            details: [],
          },
        ];

        // Call without specifying isModelStoreContext
        const result = inferenceDataSourceRestrictionsUtil.applyInferenceDataSourceRestrictions(items, clusterId);

        // Git should be allowed (not in model store context)
        expect(result[0].disabled).toBeUndefined();
      });
    });
  });
});
